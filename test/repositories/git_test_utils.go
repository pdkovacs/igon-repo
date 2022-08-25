package repositories

import (
	"fmt"
	"os"
	"strings"
	"time"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories"
	"igo-repo/test/common"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type GitTestRepo struct {
	*repositories.GitRepository
}

func NewGitTestRepo(location string, logger zerolog.Logger) *GitTestRepo {
	gitRepo, gitRepoErr := repositories.NewGitRepository(location, true, logging.CreateUnitLogger(logger, "test-git-repository"))
	if gitRepoErr != nil {
		panic(gitRepoErr)
	}
	return &GitTestRepo{
		GitRepository: gitRepo,
	}
}

func (repo *GitTestRepo) getGitStatus() (string, error) {
	out, err := repo.ExecuteGitCommand([]string{"status"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (repo *GitTestRepo) GetCurrentCommit() (string, error) {
	out, err := repo.ExecuteGitCommand([]string{"rev-parse", "HEAD"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (repo *GitTestRepo) AssertGitCleanStatus(s *suite.Suite) {
	status, err := repo.getGitStatus()
	s.NoError(err)
	s.Contains(status, cleanStatusMessageTail)
}

func (repo *GitTestRepo) GetIconfiles() ([]string, error) {
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

type GitTestSuite struct {
	suite.Suite
	repo *GitTestRepo
}

func (s *GitTestSuite) BeforeTest(suiteName, testName string) {
	s.repo = NewGitTestRepo(common.GetTestConfig().IconDataLocationGit, logging.CreateRootLogger(logging.DebugLevel))
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)
}

func (s GitTestSuite) getCurrentCommit() (string, error) {
	return s.repo.GetCurrentCommit()
}

const cleanStatusMessageTail = "nothing to commit, working tree clean"

func (s GitTestSuite) assertGitCleanStatus() {
	s.repo.AssertGitCleanStatus(&s.Suite)
}

func (s GitTestSuite) assertFileInRepo(iconName string, iconfile domain.Iconfile) {
	filePath := s.repo.GetAbsolutePathToIconfile(iconName, iconfile.IconfileDescriptor)
	fi, statErr := os.Stat(filePath)
	s.NoError(statErr)
	timeFileBorn := fi.ModTime().Unix()
	const SECONDS_IN_MILLIES = 1000
	var time3secsBackInThePast = (time.Now().Unix() - 3*SECONDS_IN_MILLIES)
	s.Greater(timeFileBorn, time3secsBackInThePast)
}

func (s GitTestSuite) assertFileNotInRepo(iconName string, iconfile domain.Iconfile) {
	var filePath = s.repo.GetAbsolutePathToIconfile(iconName, iconfile.IconfileDescriptor)
	_, statErr := os.Stat(filePath)
	s.True(os.IsNotExist(statErr))
}
