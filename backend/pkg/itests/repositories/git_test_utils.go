package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

var homeTmpDir = filepath.Join(os.Getenv("HOME"), "tmp")
var testTmpDir = filepath.Join(homeTmpDir, "tmp-icon-repo-test")
var repoDir = filepath.Join(testTmpDir, strconv.Itoa(os.Getpid()))

func createTestGitRepo() (*repositories.GitRepository, error) {
	var gitRepo *repositories.GitRepository
	var err error

	err = rmdirMaybe(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create test git repo: %w", err)
	}

	gitRepo, err = repositories.NewGitRepo(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create test git repo: %w", err)
	}

	return gitRepo, nil
}

func deleteTestGitRepo() error {
	err := os.RemoveAll(repoDir)
	if err != nil {
		return fmt.Errorf("failed to delete test git repo %s: %w", repoDir, err)
	}
	return nil
}

func rmdirMaybe(dir string) error {
	var err error
	var fi os.FileInfo

	fi, err = os.Stat(repoDir)
	if err != nil {
		if os.IsNotExist(err) {
			// nothing to do here
			return nil
		}
		return fmt.Errorf("failed to stat %s: %w", dir, err)
	}

	if fi.Mode().IsRegular() {
		return fmt.Errorf("file exists, but it is not a directory: %s", dir)
	}

	err = os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}

	return nil
}

type gitTestSuite struct {
	suite.Suite
	Repo repositories.GitRepository
}

func (s *gitTestSuite) BeforeTest(suiteName, testName string) {
	gitRepo, err := createTestGitRepo()
	if err != nil {
		panic(err)
	}
	s.Repo = *gitRepo
}

func (s gitTestSuite) AfterTest(suiteName, testName string) {
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)
	err := deleteTestGitRepo()
	if err != nil {
		panic(err)
	}
}

func (s gitTestSuite) getCurrentCommit() (string, error) {
	out, err := s.Repo.ExecuteGitCommand([]string{"rev-parse", "HEAD"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (s gitTestSuite) getGitStatus() (string, error) {
	out, err := s.Repo.ExecuteGitCommand([]string{"status"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

const cleanStatusMessageTail = "nothing to commit, working tree clean"

func (s gitTestSuite) assertGitCleanStatus() {
	status, err := s.getGitStatus()
	s.NoError(err)
	s.Contains(status, cleanStatusMessageTail)
}

func (s gitTestSuite) assertFileInRepo(iconName string, iconfile domain.Iconfile) {
	filePath := s.Repo.GetPathToIconfile(iconName, iconfile)
	fi, statErr := os.Stat(filePath)
	s.NoError(statErr)
	timeFileBorn := fi.ModTime().Unix()
	const SECONDS_IN_MILLIES = 1000
	var time3secsBackInThePast = (time.Now().Unix() - 3*SECONDS_IN_MILLIES)
	s.Greater(timeFileBorn, time3secsBackInThePast)
}

func (s gitTestSuite) assertFileNotInRepo(iconName string, iconfile domain.Iconfile) {
	var filePath = s.Repo.GetPathToIconfile(iconName, iconfile)
	_, statErr := os.Stat(filePath)
	s.True(os.IsNotExist(statErr))
}
