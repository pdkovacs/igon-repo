package git

import (
	"context"
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/test/test_commons"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type gitlabRepoTestSuite struct {
	suite.Suite
	ctx     context.Context
	t       *testing.T
	gitRepo *git.Gitlab
}

func TestGitlabRepoTestSuite(t *testing.T) {
	if len(os.Getenv("LOCAL_GIT_ONLY")) > 0 {
		return
	}
	suite.Run(t, &gitlabRepoTestSuite{ctx: context.Background(), t: t})
}

func (testSuite *gitlabRepoTestSuite) BeforeTest(suiteName string, testName string) {
	defaultTestConfig := test_commons.GetTestConfig()
	conf := test_commons.CloneConfig(defaultTestConfig)
	var createClientErr error
	testSuite.gitRepo, createClientErr = NewGitlabTestRepoClient(&conf)
	if createClientErr != nil {
		testSuite.FailNow("", "%v", createClientErr)
	}
	createRepoErr := testSuite.gitRepo.ResetRepository(testSuite.ctx)
	if createRepoErr != nil {
		testSuite.FailNow("", "%v", createRepoErr)
	}
}

func (testSuite *gitlabRepoTestSuite) AfterTest(suiteName string, testName string) {
	testSuite.gitRepo.DeleteRepository(testSuite.ctx)
}

func (testSuite *gitlabRepoTestSuite) TestAddIconfile() {
	var err error
	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]
	err = testSuite.gitRepo.AddIconfile(testSuite.ctx, icon.Name, iconfile, icon.ModifiedBy)
	testSuite.NoError(err)

	// var sha1 string
	// sha1, err = testSuite.GetStateID()
	testSuite.NoError(err)
	// testSuite.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	// testSuite.assertGitCleanStatus()
	// testSuite.assertFileInRepo(icon.Name, iconfile)
}
