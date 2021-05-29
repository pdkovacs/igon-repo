package repositories

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
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

func (g GitRepository) getPathComponents1(iconName string, iconfile domain.Iconfile) iconfilePathComponents {
	return g.getPathComponents(
		iconName,
		iconfile.Format,
		iconfile.Size,
	)
}

func (g GitRepository) GetPathToIconfile(iconName string, iconfile domain.Iconfile) string {
	return g.getPathComponents1(iconName, iconfile).pathToIconfile
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
	pathComponents := g.getPathComponents1(iconName, iconfile)
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

func NewGitRepo(location string) (*GitRepository, error) {
	var err error
	var out string

	var cmds []auxiliaries.ExecCmdParams = []auxiliaries.ExecCmdParams{
		{Name: "rm", Args: []string{"-rf", location}, Opts: nil},
		{Name: "mkdir", Args: []string{"-p", location}, Opts: nil},
		{Name: "git", Args: []string{"init"}, Opts: &auxiliaries.CmdOpts{Cwd: location}},
		{Name: "git", Args: []string{"config", "user.name", "Icon Repo Server"}, Opts: &auxiliaries.CmdOpts{Cwd: location}},
		{Name: "git", Args: []string{"config", "user.email", "IconRepoServer@UIToolBox"}, Opts: &auxiliaries.CmdOpts{Cwd: location}},
	}

	for _, cmd := range cmds {
		out, err = auxiliaries.ExecuteCommand(cmd)
		println(out)
		if err != nil {
			return nil, fmt.Errorf("failed to create git repo at %s: %w", location, err)
		}
	}

	return &GitRepository{
		Location: location,
	}, nil
}
