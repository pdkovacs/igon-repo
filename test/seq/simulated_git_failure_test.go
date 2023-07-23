package sequential_tests

import (
	"iconrepo/internal/repositories/blobstore/git"
	api_tests "iconrepo/test/api"
	blobstore_tests "iconrepo/test/repositories/blobstore"
	"iconrepo/test/test_commons"
	"iconrepo/test/testdata"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type gitTests struct {
	blobstore_tests.BlobstoreTestSuite
}

func TestGitTestSuite(t *testing.T) {
	for _, repoController := range blobstore_tests.BlobstoreProvidersToTest() {
		suite.Run(t, &gitTests{BlobstoreTestSuite: blobstore_tests.BlobstoreTestSuite{RepoController: repoController, TestSequenceId: "simulated_git_failuer"}})
	}
}

func (s *gitTests) TestRemainsConsistentAfterAddingIconfileFails() {
	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	now := time.Now()
	errorWhenAddingFirstIconFile := s.RepoController.AddIconfile(icon.Name, iconfile1, icon.ModifiedBy)
	s.NoError(errorWhenAddingFirstIconFile)

	os.Setenv(git.SimulateGitCommitFailureEnvvarName, "true")

	lastGoodSha1, errorWhenGettingLastGoodSha1 := s.GetStateID()
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(lastGoodSha1))
	s.NoError(errorWhenGettingLastGoodSha1)
	errorAddingSecondIconfile := s.RepoController.AddIconfile(icon.Name, iconfile2, icon.ModifiedBy)
	s.Error(errorAddingSecondIconfile)
	postSha1, errorWhenGettingPostSha1 := s.GetStateID()
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(postSha1))
	s.NoError(errorWhenGettingPostSha1)
	s.Equal(lastGoodSha1, postSha1)
	s.AssertBlobstoreCleanStatus()
	s.AssertFileInBlobstore(icon.Name, iconfile1, now)
	s.AssertFileNotInBlobstore(icon.Name, iconfile2)
}

type iconCreateTests struct {
	api_tests.IconTestSuite
}

func TestIconCreateTestSuite(t *testing.T) {
	for _, iconTestSuite := range api_tests.IconTestSuites("sequential_tests") {
		suite.Run(t, &iconCreateTests{IconTestSuite: iconTestSuite})
	}
}

func (s *iconCreateTests) TestRollbackToLastConsistentStateOnError() {
	moreDataIn, _ := testdata.Get()
	dataIn, dataOut := testdata.Get()

	session := s.Client.MustLoginSetAllPerms()
	session.MustAddTestData(dataIn)

	lastStableSHA1, beforeIncidentGitErr := s.TestBlobstoreController.GetStateID()
	s.NoError(beforeIncidentGitErr)

	os.Setenv(git.SimulateGitCommitFailureEnvvarName, "true")

	statusCode, _, _ := session.CreateIcon(moreDataIn[1].Name, moreDataIn[1].Iconfiles[0].Content)
	s.Equal(http.StatusConflict, statusCode)

	afterIncidentSHA1, afterIncidentGitErr := s.TestBlobstoreController.GetStateID()
	s.NoError(afterIncidentGitErr)

	s.Equal(lastStableSHA1, afterIncidentSHA1)

	iconDescriptors, describeError := session.DescribeAllIcons(s.Ctx)
	s.NoError(describeError)
	s.AssertResponseIconSetsEqual(dataOut, iconDescriptors)

	s.AssertEndState()
}
