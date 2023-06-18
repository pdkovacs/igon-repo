package gitrepo

import (
	"fmt"
	"os"
	"strings"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"

	"github.com/rs/zerolog"
)

type Local struct {
	Location  string
	Logger    zerolog.Logger
	FilePaths filePaths
}

func (repo Local) String() string {
	return fmt.Sprintf("Local git repository at %s", repo.Location)
}

func NewLocalGitRepository(location string) Local {
	git := Local{
		Location:  location,
		FilePaths: NewGitFilePaths(location),
	}
	return git
}

var SimulateGitCommitFailureEnvvarName = "GIT_COMMIT_FAIL_INTRUSIVE_TEST"
var gitCommitFailureTestCommand = "procyon lotor"

const filesAddedSuccessMessage = "icon file(s) added"
const filesDeletedSuccessMessage = "all file(s) for icon \"%s\" deleted:\n\n%s"
const fileDeleteSuccessMessage = "iconfile for icon \"%s\" deleted:\n\n%s"

const cleanStatusMessageTail = "nothing to commit, working tree clean"

func (repo Local) Create() error {
	return repo.initMaybe()
}

func (repo Local) Delete() error {
	return os.RemoveAll(repo.Location)
}

func (repo Local) GetAbsolutePathToIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor) string {
	return repo.FilePaths.GetAbsolutePathToIconfile(iconName, iconfileDescriptor)
}

func (repo Local) ExecuteGitCommand(args []string) (string, error) {
	return config.ExecuteCommand(config.ExecCmdParams{
		Name: "git",
		Args: args,
		Opts: &config.CmdOpts{Cwd: repo.Location},
	}, repo.Logger)
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
	if os.Getenv(SimulateGitCommitFailureEnvvarName) == "true" {
		return gitCommitFailureTestCommand
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

func (repo Local) rollback() {
	for _, rollbackCmd := range rollbackCommands {
		_, _ = repo.ExecuteGitCommand(rollbackCmd)
	}
}

func (repo Local) executeIconfileJob(iconfileOperation func() ([]string, error), messages gitJobTextProvider, userName string) error {
	logger := logging.CreateMethodLogger(repo.Logger, fmt.Sprintf("git: %s", messages.logContext))

	var out string
	var err error
	var iconfilePathsInRepo []string

	defer func() {
		if err != nil {
			logger.Debug().Err(err).Str("out", out).Msg("failed GIT operation")
			repo.rollback()
		} else {
			logger.Debug().Msg("Success")
		}
	}()

	iconfilePathsInRepo, err = iconfileOperation()
	if err != nil {
		return fmt.Errorf("failed iconfile operation: %w", err)
	}
	out, err = repo.ExecuteGitCommand([]string{"add", "-A"})
	if err != nil {
		return fmt.Errorf("failed to add files to index: %w -> %s", err, out)
	}

	commitMessage := messages.getCommitMessage(iconfilePathsInRepo)
	out, err = repo.ExecuteGitCommand(commit(commitMessage, userName))
	if err != nil {
		return fmt.Errorf("failed to commit: %w -> %s", err, out)
	}

	return err
}

func (repo Local) createIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) (string, error) {
	pathComponents := repo.FilePaths.getPathComponents(iconName, iconfile.IconfileDescriptor)
	var err error

	logOp := func(opmsg string) string {
		repo.Logger.Debug().Str("operation", opmsg).Msg("operation starting")
		return opmsg
	}

	operationMsg := logOp(fmt.Sprintf("create directory %s", pathComponents.pathToFormatDir))
	err = os.MkdirAll(pathComponents.pathToFormatDir, 0700)
	if err == nil {
		operationMsg = logOp(fmt.Sprintf("create directory %s", pathComponents.pathToSizeDir))
		err = os.MkdirAll(pathComponents.pathToSizeDir, 0700)
		if err == nil {
			operationMsg = logOp(fmt.Sprintf("write file %s", pathComponents.pathToIconfile))
			err = os.WriteFile(pathComponents.pathToIconfile, iconfile.Content, 0700)
		}
	}
	if err != nil {
		err = fmt.Errorf("%s: %w", operationMsg, err)
	}
	return pathComponents.pathToIconfileInRepo, err
}

func (repo Local) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error {
	iconfileOperation := func() ([]string, error) {
		pathToIconfileInRepo, err := repo.createIconfile(iconName, iconfile, modifiedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to create iconfile %v for %s: %w", iconfile, iconName, err)
		}
		return []string{pathToIconfileInRepo}, nil
	}

	jobTextProvider := gitJobTextProvider{
		"add icon file",
		defaultCommitMessageProvider(filesAddedSuccessMessage),
	}

	var err error
	config.Enqueue(func() {
		err = repo.executeIconfileJob(iconfileOperation, jobTextProvider, modifiedBy)
	})

	if err != nil {
		return fmt.Errorf("failed to add iconfile %v for %s to git repository: %w", iconfile, iconName, err)
	}
	return nil
}

func (repo Local) GetIconfile(iconName string, iconfileDesc domain.IconfileDescriptor) ([]byte, error) {
	pathToFile := repo.GetAbsolutePathToIconfile(iconName, iconfileDesc)
	bytes, err := os.ReadFile(pathToFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s from local git repo: %w", pathToFile, err)
	}
	return bytes, nil
}

func (repo Local) deleteIconfileFile(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error) {
	pathCompos := repo.FilePaths.getPathComponents(iconName, iconfileDesc)
	removeFileErr := os.Remove(pathCompos.pathToIconfile)
	if removeFileErr != nil {
		if removeFileErr == os.ErrNotExist {
			return "", fmt.Errorf("failed to remove iconfile %v for icon %s: %w", iconfileDesc, iconName, domain.ErrIconNotFound)
		}
		return "", fmt.Errorf("failed to remove iconfile %v for icon %s: %w", iconfileDesc, iconName, removeFileErr)
	}
	return pathCompos.pathToIconfileInRepo, nil
}

func (repo Local) DeleteIcon(iconDesc domain.IconDescriptor, modifiedBy authn.UserID) error {
	iconfileOperation := func() ([]string, error) {
		var opError error
		var fileList []string

		for _, ifDesc := range iconDesc.Iconfiles {
			filePath, deletionError := repo.deleteIconfileFile(iconDesc.Name, ifDesc)
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
			return fmt.Sprintf(filesDeletedSuccessMessage, iconDesc.Name, fileListAsText(fileList))
		},
	}

	var err error
	config.Enqueue(func() {
		err = repo.executeIconfileJob(iconfileOperation, jobTextProvider, modifiedBy.String())
	})

	if err != nil {
		return fmt.Errorf("failed to remove icon %s from git repository: %w", iconDesc.Name, err)
	}
	return nil
}

func (repo Local) DeleteIconfile(iconName string, iconfileDesc domain.IconfileDescriptor, modifiedBy authn.UserID) error {
	iconfileOperation := func() ([]string, error) {
		filePath, deletionError := repo.deleteIconfileFile(iconName, iconfileDesc)
		return []string{filePath}, deletionError
	}

	jobTextProvider := gitJobTextProvider{
		fmt.Sprintf("delete iconfile %v for icon \"%s\"", iconfileDesc, iconName),
		func(fileList []string) string {
			return fmt.Sprintf(fileDeleteSuccessMessage, iconName, fileListAsText(fileList))
		},
	}

	var err error
	config.Enqueue(func() {
		err = repo.executeIconfileJob(iconfileOperation, jobTextProvider, modifiedBy.String())
	})

	if err != nil {
		return fmt.Errorf("failed to remove iconfile %v of \"%s\" from git repository: %w", iconfileDesc, iconName, err)
	}
	return nil
}

func (repo Local) CheckStatus() (bool, error) {
	out, err := repo.ExecuteGitCommand([]string{"status"})
	if err != nil {
		return false, fmt.Errorf("failed to get current git commit: %w", err)
	}
	status := strings.TrimSpace(out)
	return strings.Contains(status, cleanStatusMessageTail), nil
}

func (repo Local) GetStateID() (string, error) {
	out, err := repo.ExecuteGitCommand([]string{"rev-parse", "HEAD"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (repo Local) GetIconfiles() ([]string, error) {
	output, err := repo.ExecuteGitCommand([]string{"ls-tree", "-r", "HEAD", "--name-only"})
	if err != nil {
		return nil, err
	}

	fileList := []string{}
	outputLines := strings.Split(output, config.LineBreak)
	for _, line := range outputLines {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) > 0 {
			fileList = append(fileList, trimmedLine)
		}
	}
	return fileList, nil
}

// GetCommitIDFor returns the commit ID of the iconfile specified by the method paramters.
// Return empty string in case the file doesn't exist in the repository
func (repo Local) GetCommitIDFor(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error) {
	printCommitIDArgs := []string{"log", "-n", "1", "--pretty=format:%H", "--", repo.FilePaths.GetPathToIconfileInRepo(iconName, iconfileDesc)}
	output, execErr := repo.ExecuteGitCommand(printCommitIDArgs)
	if execErr != nil {
		return "", fmt.Errorf("failed to execute command to get last commit modifying %s::%s: %w", iconName, iconfileDesc.String(), execErr)
	}
	return output, nil
}

func (repo Local) GetCommitMetadata(commitId string) (CommitMetadata, error) {
	logger := logging.CreateMethodLogger(repo.Logger, fmt.Sprintf("git: GetCommitMetadata: %s", commitId))

	printCommitMetadataArgs := []string{"show", "--quiet", "--format=fuller", "--date=format:%Y-%m-%dT%H:%M:%S%z"}
	output, execErr := repo.ExecuteGitCommand(printCommitMetadataArgs)
	if execErr != nil {
		return CommitMetadata{}, fmt.Errorf("failed to get metadata from repo for commit %s: %w", commitId, execErr)
	}
	logger.Debug().Str("meta-data", output).Msg("raw metadata extracted")
	commitMetadata, parseErr := parseLocalCommitMetadata(output)
	if parseErr != nil {
		return commitMetadata, fmt.Errorf("failed to parse metadata from commit %s: %w", commitId, parseErr)
	}
	return commitMetadata, nil
}

func (repo Local) createInitializeGitRepo() error {
	var err error
	var out string

	var cmds []config.ExecCmdParams = []config.ExecCmdParams{
		{Name: "rm", Args: []string{"-rf", repo.Location}, Opts: nil},
		{Name: "mkdir", Args: []string{"-p", repo.Location}, Opts: nil},
		{Name: "git", Args: []string{"init"}, Opts: &config.CmdOpts{Cwd: repo.Location}},
		{Name: "git", Args: []string{"config", "user.name", "Icon Repo Server"}, Opts: &config.CmdOpts{Cwd: repo.Location}},
		{Name: "git", Args: []string{"config", "user.email", "IconRepoServer@UIToolBox"}, Opts: &config.CmdOpts{Cwd: repo.Location}},
	}

	for _, cmd := range cmds {
		out, err = config.ExecuteCommand(cmd, repo.Logger)
		println(out)
		if err != nil {
			return fmt.Errorf("failed to create git repo at %s: %w", repo.Location, err)
		}
	}

	return nil
}

func (repo Local) locationHasRepo() bool {
	if GitRepoLocationExists(repo.Location) {
		testCommand := config.ExecCmdParams{Name: "git", Args: []string{"init"}, Opts: &config.CmdOpts{Cwd: repo.Location}}
		outOrErr, err := config.ExecuteCommand(testCommand, repo.Logger)
		if err != nil {
			if strings.Contains(outOrErr, "not a git repository") { // TODO: Is it really possible to get this error message here?
				return false
			}
			panic(err)
		}
		return true
	}
	return false
}

// Init initializes the Git repository if it already doesn't exist
func (repo Local) initMaybe() error {
	if !repo.locationHasRepo() {
		return repo.createInitializeGitRepo()
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
