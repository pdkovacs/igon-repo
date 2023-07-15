package git

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

func NewLocalGitTestRepo(conf config.Options) git.Local {
	return git.NewLocalGitRepository(conf.LocalGitRepo)
}

func NewGitlabTestRepoClient(conf config.Options) git.Gitlab {
	gitlab, err := git.NewGitlabRepositoryClient(
		conf.GitlabNamespacePath,
		conf.GitlabProjectPath,
		conf.GitlabMainBranch,
		conf.GitlabAccessToken,
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
	GetCommitMetadata(commitId string) (git.CommitMetadata, error)
}

type GitTestRepo interface {
	repositories.BlobstoreRepository
	gitRepoManagement
}

func GitProvidersToTest() []GitTestRepo {
	if len(os.Getenv("LOCAL_GIT_ONLY")) > 0 {
		return []GitTestRepo{git.Local{}}
	}
	return []GitTestRepo{
		git.Local{},
		git.Gitlab{},
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
	os.Unsetenv(git.SimulateGitCommitFailureEnvvarName)
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
	case git.Local:
		s.Repo = NewLocalGitTestRepo(conf)
	case git.Gitlab:
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

func (s *GitTestSuite) GetStateID() (string, error) {
	return s.Repo.GetStateID()
}

func (s *GitTestSuite) AssertGitCleanStatus() {
	status, err := s.Repo.CheckStatus()
	s.NoError(err)
	s.True(status)
}

func (s *GitTestSuite) AssertFileInRepo(iconName string, iconfile domain.Iconfile, timeBeforeCommit time.Time) {
	commitID, getCommitIDErr := s.Repo.GetCommitIDFor(iconName, iconfile.IconfileDescriptor)
	s.NoError(getCommitIDErr)
	s.Greater(len(commitID), 0)
	meta, getMetaErr := s.Repo.GetCommitMetadata(commitID)
	s.NoError(getMetaErr)
	s.Greater(meta.CommitDate, timeBeforeCommit.Add(-time.Duration(1_000)*time.Millisecond))
}

func (s *GitTestSuite) AssertFileNotInRepo(iconName string, iconfile domain.Iconfile) {
	commitId, err := s.Repo.GetCommitIDFor(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal("", commitId)
}
