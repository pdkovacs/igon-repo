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
var repoBaseDir = filepath.Join(testTmpDir, strconv.Itoa(os.Getpid()))

func (s *GitTestSuite) RecreateInitTestGitRepo() {
	var err error

	repoDir := filepath.Join(repoBaseDir, s.repo.Location)

	err = rmdirMaybe(repoDir)
	if err != nil {
		panic(err)
	}

	err = s.repo.InitMaybe()
	if err != nil {
		panic(err)
	}
}

func rmdirMaybe(dir string) error {
	if repositories.GitRepoLocationExists(dir) {
		err := os.RemoveAll(dir)
		if err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", dir, err)
		}
		return nil
	}
	return nil
}

func DeleteTestGitRepo(repoLocation string) error {
	return rmdirMaybe(repoLocation)
}

type GitTestSuite struct {
	suite.Suite
	repo repositories.GitRepository
}

func (s *GitTestSuite) SetupSuite() {
	s.repo = repositories.GitRepository{Location: "itest-repositories"}
}

func (s *GitTestSuite) BeforeTest(suiteName, testName string) {
	err := DeleteTestGitRepo(s.repo.Location)
	if err != nil {
		panic(err)
	}
	s.RecreateInitTestGitRepo()
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)
}

func (s GitTestSuite) getCurrentCommit() (string, error) {
	out, err := s.repo.ExecuteGitCommand([]string{"rev-parse", "HEAD"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

const cleanStatusMessageTail = "nothing to commit, working tree clean"

func (s GitTestSuite) assertGitCleanStatus() {
	AssertGitCleanStatus(&s.Suite, &s.repo)
}

func (s GitTestSuite) assertFileInRepo(iconName string, iconfile domain.Iconfile) {
	filePath := s.repo.GetPathToIconfile(iconName, iconfile)
	fi, statErr := os.Stat(filePath)
	s.NoError(statErr)
	timeFileBorn := fi.ModTime().Unix()
	const SECONDS_IN_MILLIES = 1000
	var time3secsBackInThePast = (time.Now().Unix() - 3*SECONDS_IN_MILLIES)
	s.Greater(timeFileBorn, time3secsBackInThePast)
}

func (s GitTestSuite) assertFileNotInRepo(iconName string, iconfile domain.Iconfile) {
	var filePath = s.repo.GetPathToIconfile(iconName, iconfile)
	_, statErr := os.Stat(filePath)
	s.True(os.IsNotExist(statErr))
}

func getGitStatus(repo *repositories.GitRepository) (string, error) {
	out, err := repo.ExecuteGitCommand([]string{"status"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func AssertGitCleanStatus(s *suite.Suite, repo *repositories.GitRepository) {
	status, err := getGitStatus(repo)
	s.NoError(err)
	s.Contains(status, cleanStatusMessageTail)
}
