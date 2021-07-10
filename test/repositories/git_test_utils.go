package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pdkovacs/igo-repo/config"
	"github.com/pdkovacs/igo-repo/domain"
	"github.com/pdkovacs/igo-repo/repositories"
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

type GitTestRepo struct {
	repositories.GitRepository
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
	repo GitTestRepo
}

func (s *GitTestSuite) SetupSuite() {
	s.repo = GitTestRepo{
		repositories.GitRepository{Location: "itest-repositories"},
	}
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
