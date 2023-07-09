package gitlab_tests

import (
	"fmt"
	"iconrepo/internal/logging"
	"iconrepo/internal/repositories/gitrepo"
	"iconrepo/test/repositories/git_tests"
	"iconrepo/test/test_commons"
	"testing"

	"github.com/stretchr/testify/suite"
)

const testSequenceId = "gitlabtests"

var gitlabRepoTestLogger = logging.CreateUnitLogger(logging.Get(), "repositories.gitlabRepoTestSuite")

type gitlabRepoTestSuite struct {
	suite.Suite
	t       *testing.T
	gitRepo gitrepo.Gitlab
}

func TestGitlabRepoTestSuite(t *testing.T) {
	suite.Run(t, &gitlabRepoTestSuite{t: t})
}

func (testSuite *gitlabRepoTestSuite) BeforeTest(suiteName string, testName string) {
	defaultTestConfig := test_commons.GetTestConfig()
	conf := test_commons.CloneConfig(defaultTestConfig)
	conf.LocalGitRepo = fmt.Sprintf("%s_%s", conf.LocalGitRepo, testSequenceId)
	conf.GitlabProjectPath = fmt.Sprintf("%s_%s", defaultTestConfig.GitlabProjectPath, testSequenceId)
	conf.GitlabAccessToken = git_tests.GitTestGitlabAPIToken()
	conf.GitlabNamespacePath = "testing-with-repositories"

	repo, err := gitrepo.NewGitlabRepositoryClient(
		conf.GitlabNamespacePath,
		conf.GitlabProjectPath+"_"+testSequenceId,
		conf.GitlabMainBranch,
		conf.GitlabAccessToken,
	)
	if err != nil {
		panic(err)
	}
	git_tests.MustResetTestGitRepo(repo)
	testSuite.gitRepo = repo
}

func (testSuite *gitlabRepoTestSuite) AfterTest(suiteName string, testName string) {
	testSuite.gitRepo.Delete()
}

func (testSuite *gitlabRepoTestSuite) TestAddIconfile() {
	var err error
	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]
	err = testSuite.gitRepo.AddIconfile(icon.Name, iconfile, icon.ModifiedBy)
	testSuite.NoError(err)

	// var sha1 string
	// sha1, err = testSuite.GetStateID()
	testSuite.NoError(err)
	// testSuite.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	// testSuite.assertGitCleanStatus()
	// testSuite.assertFileInRepo(icon.Name, iconfile)
}
