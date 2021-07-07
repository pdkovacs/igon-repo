package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authn"
	log "github.com/sirupsen/logrus"
)

type GitRepository struct {
	Location string
}

var IntrusiveGitTestEnvvarName = "GIT_COMMIT_FAIL_INTRUSIVE_TEST"
var intrusiveGitTestCommand = "procyon lotor"

type iconfilePathComponents struct {
	pathToFormatDir      string
	pathToSizeDir        string
	pathToIconfile       string
	pathToIconfileInRepo string
}

func getFileName(iconName string, format string, size string) string {
	return fmt.Sprintf("%s@%s.%s", iconName, size, format)
}

func (g GitRepository) getPathComponents(iconName string, format string, size string) iconfilePathComponents {
	fileName := getFileName(iconName, format, size)
	pathToFormatDir := filepath.Join(g.Location, format)
	pathToSizeDir := filepath.Join(pathToFormatDir, size)
	pathToIconfile := filepath.Join(pathToSizeDir, fileName)
	pathToIconfileInRepo := filepath.Join(format, filepath.Join(size, fileName))
	return iconfilePathComponents{
		pathToFormatDir,
		pathToSizeDir,
		pathToIconfile,
		pathToIconfileInRepo,
	}
}

func (g GitRepository) getPathComponents1(iconName string, iconfile domain.IconfileDescriptor) iconfilePathComponents {
	return g.getPathComponents(
		iconName,
		iconfile.Format,
		iconfile.Size,
	)
}

func (g GitRepository) GetAbsolutePathToIconfile(iconName string, iconfile domain.IconfileDescriptor) string {
	return g.getPathComponents1(iconName, iconfile).pathToIconfile
}

func (g GitRepository) GetPathToIconfileInRepos(iconName string, iconfile domain.IconfileDescriptor) string {
	return g.getPathComponents1(iconName, iconfile).pathToIconfileInRepo
}

func (g GitRepository) ExecuteGitCommand(args []string) (string, error) {
	return auxiliaries.ExecuteCommand(auxiliaries.ExecCmdParams{
		Name: "git",
		Args: args,
		Opts: &auxiliaries.CmdOpts{Cwd: g.Location},
	})
}

type getCommitMessageFn func(filelist []string) string

type gitJobTextProvider struct {
	logContext       string
	getCommitMessage getCommitMessageFn
}

func addPathInRepo(fileListText string, pathInRepo string) string {
	separator := ""
	if len(fileListText) > 0 {
		separator = "\n"
	}
	return fileListText + separator + pathInRepo
}

func fileListAsText(fileList []string) string {
	fileListText := ""
	for _, item := range fileList {
		fileListText = addPathInRepo(fileListText, item)
	}
	return fileListText
}

func defaultCommitMessageProvider(messageBase string) getCommitMessageFn {
	return func(fileList []string) string {
		return fileListAsText(fileList) + " " + messageBase
	}
}

func getCommitCommand() string {
	if os.Getenv(IntrusiveGitTestEnvvarName) == "true" {
		return intrusiveGitTestCommand
	} else {
		return "commit"
	}
}

func commit(messageBase string, userName string) []string {
	return []string{
		getCommitCommand(),
		"-m", messageBase + " by " + userName,
		fmt.Sprintf("--author=%s@IconRepoServer <%s>", userName, userName),
	}
}

var rollbackCommands = [][]string{
	{"reset", "--hard", "HEAD"},
	{"clean", "-qfdx"},
}

func (g GitRepository) rollback() {
	for _, rollbackCmd := range rollbackCommands {
		_, _ = g.ExecuteGitCommand(rollbackCmd)
	}
}

func (g GitRepository) createIconfileJob(iconfileOperation func() ([]string, error), messages gitJobTextProvider, userName string) error {
	logger := log.WithField("prefix", fmt.Sprintf("git: %s", messages.logContext))

	var err error
	var iconfilePathsInRepo []string

	defer func() {
		if err != nil {
			logger.Errorf("failed to create iconfile: %v", err)
			g.rollback()
		} else {
			logger.Debug("Success")
		}
	}()

	iconfilePathsInRepo, err = iconfileOperation()
	if err != nil {
		return fmt.Errorf("failed iconfile operation: %w", err)
	}
	_, err = g.ExecuteGitCommand([]string{"add", "-A"})

	commitMessage := messages.getCommitMessage(iconfilePathsInRepo)
	_, err = g.ExecuteGitCommand(commit(commitMessage, userName))
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return err
}

func (g GitRepository) createIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) (string, error) {
	pathComponents := g.getPathComponents1(iconName, iconfile.IconfileDescriptor)
	var err error
	err = os.MkdirAll(pathComponents.pathToFormatDir, 0700)
	if err == nil {
		err = os.MkdirAll(pathComponents.pathToSizeDir, 0700)
		if err == nil {
			err = os.WriteFile(pathComponents.pathToIconfile, iconfile.Content, 0700)
		}
	}
	return pathComponents.pathToIconfileInRepo, err
}

func (g GitRepository) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error {
	iconfileOperation := func() ([]string, error) {
		pathToIconfileInRepo, err := g.createIconfile(iconName, iconfile, modifiedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to create iconfile %v for %s: %w", iconfile, iconName, err)
		}
		return []string{pathToIconfileInRepo}, nil
	}

	jobTextProvider := gitJobTextProvider{
		"add icon file",
		defaultCommitMessageProvider("icon file(s) added"),
	}

	var err error
	auxiliaries.Enqueue(func() {
		err = g.createIconfileJob(iconfileOperation, jobTextProvider, modifiedBy)
	})

	if err != nil {
		return fmt.Errorf("failed to add iconfile %v for %s to git repository: %w", iconfile, iconName, err)
	}
	return nil
}

func (s *GitRepository) deleteIconfileFile(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error) {
	pathCompos := s.getPathComponents1(iconName, iconfileDesc)
	removeFileErr := os.Remove(pathCompos.pathToIconfile)
	if removeFileErr != nil {
		if removeFileErr == os.ErrNotExist {
			return "", fmt.Errorf("failed to remove iconfile %v for icon %s: %w", iconfileDesc, iconName, domain.ErrIconNotFound)
		}
		return "", fmt.Errorf("failed to remove iconfile %v for icon %s: %w", iconfileDesc, iconName, removeFileErr)
	}
	return pathCompos.pathToIconfileInRepo, nil
}

func (s *GitRepository) DeleteIcon(iconDesc domain.IconDescriptor, modifiedBy authn.UserID) error {
	iconfileOperation := func() ([]string, error) {
		var opError error
		var fileList []string

		for _, ifDesc := range iconDesc.Iconfiles {
			filePath, deletionError := s.deleteIconfileFile(iconDesc.Name, ifDesc)
			if deletionError != nil {
				opError = deletionError
				break
			}
			fileList = append(fileList, filePath)
		}
		return fileList, opError
	}

	jobTextProvider := gitJobTextProvider{
		fmt.Sprintf("delete all files for icon \"%s\"", iconDesc.Name),
		func(fileList []string) string {
			return fmt.Sprintf("all file(s) for icon \"%s\" deleted:\n\n%s", iconDesc.Name, fileListAsText(fileList))
		},
	}

	var err error
	auxiliaries.Enqueue(func() {
		err = s.createIconfileJob(iconfileOperation, jobTextProvider, modifiedBy.String())
	})

	if err != nil {
		return fmt.Errorf("failed to remove icon %s from git repository: %w", iconDesc.Name, err)
	}
	return nil
}

func (s *GitRepository) DeleteIconfile(iconName string, iconfileDesc domain.IconfileDescriptor, modifiedBy authn.UserID) error {
	iconfileOperation := func() ([]string, error) {
		filePath, deletionError := s.deleteIconfileFile(iconName, iconfileDesc)
		return []string{filePath}, deletionError
	}

	jobTextProvider := gitJobTextProvider{
		fmt.Sprintf("delete iconfile %v for icon \"%s\"", iconfileDesc, iconName),
		func(fileList []string) string {
			return fmt.Sprintf("iconfile for icon \"%s\" deleted:\n\n%s", iconName, fileListAsText(fileList))
		},
	}

	var err error
	auxiliaries.Enqueue(func() {
		err = s.createIconfileJob(iconfileOperation, jobTextProvider, modifiedBy.String())
	})

	if err != nil {
		return fmt.Errorf("failed to remove iconfile %v of \"%s\" from git repository: %w", iconfileDesc, iconName, err)
	}
	return nil
}

func (s *GitRepository) createInitializeGitRepo() error {
	var err error
	var out string

	var cmds []auxiliaries.ExecCmdParams = []auxiliaries.ExecCmdParams{
		{Name: "rm", Args: []string{"-rf", s.Location}, Opts: nil},
		{Name: "mkdir", Args: []string{"-p", s.Location}, Opts: nil},
		{Name: "git", Args: []string{"init"}, Opts: &auxiliaries.CmdOpts{Cwd: s.Location}},
		{Name: "git", Args: []string{"config", "user.name", "Icon Repo Server"}, Opts: &auxiliaries.CmdOpts{Cwd: s.Location}},
		{Name: "git", Args: []string{"config", "user.email", "IconRepoServer@UIToolBox"}, Opts: &auxiliaries.CmdOpts{Cwd: s.Location}},
	}

	for _, cmd := range cmds {
		out, err = auxiliaries.ExecuteCommand(cmd)
		println(out)
		if err != nil {
			return fmt.Errorf("failed to create git repo at %s: %w", s.Location, err)
		}
	}

	return nil
}

func (s *GitRepository) test() bool {
	if GitRepoLocationExists(s.Location) {
		testCommand := auxiliaries.ExecCmdParams{Name: "git", Args: []string{"init"}, Opts: &auxiliaries.CmdOpts{Cwd: s.Location}}
		outOrErr, err := auxiliaries.ExecuteCommand(testCommand)
		if err != nil {
			if strings.Contains(outOrErr, "not a git repository") {
				return false
			}
			panic(err)
		}
	}
	return false
}

// Init initializes the Git repository if it already doesn't exist
func (s *GitRepository) InitMaybe() error {
	if !s.test() {
		return s.createInitializeGitRepo()
	}
	return nil
}

func GitRepoLocationExists(location string) bool {
	var err error
	var fi os.FileInfo

	fi, err = os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			// nothing to do here
			return false
		}
		panic(err)
	}

	if fi.Mode().IsRegular() {
		panic(fmt.Errorf("file exists, but it is not a directory: %s", location))
	}

	return true
}
