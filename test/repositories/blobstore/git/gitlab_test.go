package git

import (
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/test/test_commons"
	"testing"

	"github.com/stretchr/testify/suite"
)

type gitlabRepoTestSuite struct {
	suite.Suite
	t       *testing.T
	gitRepo *git.Gitlab
}

func TestGitlabRepoTestSuite(t *testing.T) {
	suite.Run(t, &gitlabRepoTestSuite{t: t})
}

func (testSuite *gitlabRepoTestSuite) BeforeTest(suiteName string, testName string) {
	defaultTestConfig := test_commons.GetTestConfig()
	conf := test_commons.CloneConfig(defaultTestConfig)
	var createClientErr error
	testSuite.gitRepo, createClientErr = NewGitlabTestRepoClient(&conf)
	if createClientErr != nil {
		testSuite.FailNow("", "%v", createClientErr)
	}
	createRepoErr := testSuite.gitRepo.ResetRepository()
	if createRepoErr != nil {
		testSuite.FailNow("", "%v", createRepoErr)
	}
}

func (testSuite *gitlabRepoTestSuite) AfterTest(suiteName string, testName string) {
	testSuite.gitRepo.DeleteRepository()
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
