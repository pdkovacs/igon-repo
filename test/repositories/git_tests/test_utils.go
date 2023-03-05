package git_tests

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories"
	"igo-repo/internal/repositories/gitrepo"
	"igo-repo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

func NewLocalGitTestRepo(conf config.Options) gitrepo.Local {
	return gitrepo.NewLocalGitRepository(conf.LocalGitRepo, logging.CreateRootLogger(logging.DebugLevel))
}

func NewGitlabTestRepoClient(conf config.Options) gitrepo.Gitlab {
	gitlab, err := gitrepo.NewGitlabRepositoryClient(
		conf.GitlabNamespacePath,
		conf.GitlabProjectPath,
		conf.GitlabMainBranch,
		conf.GitlabAccessToken,
		logging.CreateRootLogger(logging.DebugLevel),
	)
	if err != nil {
		panic(err)
	}
	return gitlab
}

type gitRepoManagement interface {
	Delete() error
	CheckStatus() (bool, error)
	GetStateID() (string, error)
	GetIconfiles() ([]string, error)
	GetIconfile(iconName string, iconfileDesc domain.IconfileDescriptor) ([]byte, error)
	GetCommitIDFor(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error)
	GetCommitMetadata(commitId string) (gitrepo.CommitMetadata, error)
}

type GitTestRepo interface {
	repositories.GitRepository
	gitRepoManagement
}

func GitProvidersToTest() []GitTestRepo {
	return []GitTestRepo{
		gitrepo.Local{},
		gitrepo.Gitlab{},
	}
}

func MustResetTestGitRepo(repo GitTestRepo) {
	deleteRepoErr := repo.Delete()
	if deleteRepoErr != nil {
		panic(deleteRepoErr)
	}
	createRepoErr := repo.Create()
	if createRepoErr != nil {
		panic(createRepoErr)
	}
	os.Unsetenv(gitrepo.SimulateGitCommitFailureEnvvarName)
}

const gitlabAPITokenLineRegexpString = "GITLAB_ACCESS_TOKEN=?(.+)"

var gitlabAPITokenLineRegexp = regexp.MustCompile(gitlabAPITokenLineRegexpString)

func GitTestGitlabAPIToken() string {
	homeDir, homedirErr := os.UserHomeDir()
	if homedirErr != nil {
		panic(homedirErr)
	}
	content, readErr := os.ReadFile(fmt.Sprintf("%s/.icon-repo.secrets", homeDir))
	if readErr != nil {
		panic(readErr)
	}

	hasMatch := gitlabAPITokenLineRegexp.Match(content)
	if !hasMatch {
		panic(fmt.Sprintf("No match for gitlab api token pattern in content. I was looking for: %s", gitlabAPITokenLineRegexpString))
	}

	submatches := gitlabAPITokenLineRegexp.FindStringSubmatch(string(content))
	if len(submatches) < 2 {
		panic("No match for gitlab api token pattern in content")
	}
	return submatches[1]
}

type GitTestSuite struct {
	suite.Suite
	Repo GitTestRepo
}

func (s *GitTestSuite) BeforeTest(suiteName, testName string) {
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	conf.GitlabProjectPath = fmt.Sprintf("%s_%s", conf.GitlabProjectPath, "gitests")
	conf.GitlabAccessToken = GitTestGitlabAPIToken()

	switch s.Repo.(type) {
	case gitrepo.Local:
		s.Repo = NewLocalGitTestRepo(conf)
	case gitrepo.Gitlab:
		conf.GitlabNamespacePath = "testing-with-repositories"
		s.Repo = NewGitlabTestRepoClient(conf)
	case nil:
		s.Repo = NewLocalGitTestRepo(conf)
	}
	MustResetTestGitRepo(s.Repo)
}

func (s *GitTestSuite) AfterTest(suiteName, testName string) {
	s.Repo.Delete()
}

func (s GitTestSuite) GetStateID() (string, error) {
	return s.Repo.GetStateID()
}

func (s GitTestSuite) AssertGitCleanStatus() {
	status, err := s.Repo.CheckStatus()
	s.NoError(err)
	s.True(status)
}

func (s GitTestSuite) AssertFileInRepo(iconName string, iconfile domain.Iconfile, timeBeforeCommit time.Time) {
	commitID, getCommitIDErr := s.Repo.GetCommitIDFor(iconName, iconfile.IconfileDescriptor)
	s.NoError(getCommitIDErr)
	s.Greater(len(commitID), 0)
	meta, getMetaErr := s.Repo.GetCommitMetadata(commitID)
	s.NoError(getMetaErr)
	s.Greater(meta.CommitDate, timeBeforeCommit.Add(-time.Duration(1_000)*time.Millisecond))
}

func (s GitTestSuite) AssertFileNotInRepo(iconName string, iconfile domain.Iconfile) {
	commitId, err := s.Repo.GetCommitIDFor(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal("", commitId)
}
