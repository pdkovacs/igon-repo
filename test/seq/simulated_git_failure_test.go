package sequential_tests

import (
	"iconrepo/internal/repositories/gitrepo"
	api_tests "iconrepo/test/api"
	"iconrepo/test/repositories/git_tests"
	"iconrepo/test/test_commons"
	"iconrepo/test/testdata"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type gitTests struct {
	git_tests.GitTestSuite
}

func TestGitTestSuite(t *testing.T) {
	for _, repo := range git_tests.GitProvidersToTest() {
		suite.Run(t, &gitTests{GitTestSuite: git_tests.GitTestSuite{Repo: repo}})
	}
}

func (s *gitTests) TestRemainsConsistentAfterAddingIconfileFails() {
	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	now := time.Now()
	errorWhenAddingFirstIconFile := s.Repo.AddIconfile(icon.Name, iconfile1, icon.ModifiedBy)
	s.NoError(errorWhenAddingFirstIconFile)

	os.Setenv(gitrepo.SimulateGitCommitFailureEnvvarName, "true")

	lastGoodSha1, errorWhenGettingLastGoodSha1 := s.GetStateID()
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(lastGoodSha1))
	s.NoError(errorWhenGettingLastGoodSha1)
	errorAddingSecondIconfile := s.Repo.AddIconfile(icon.Name, iconfile2, icon.ModifiedBy)
	s.Error(errorAddingSecondIconfile)
	postSha1, errorWhenGettingPostSha1 := s.GetStateID()
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(postSha1))
	s.NoError(errorWhenGettingPostSha1)
	s.Equal(lastGoodSha1, postSha1)
	s.AssertGitCleanStatus()
	s.AssertFileInRepo(icon.Name, iconfile1, now)
	s.AssertFileNotInRepo(icon.Name, iconfile2)
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

	lastStableSHA1, beforeIncidentGitErr := s.TestGitRepo.GetStateID()
	s.NoError(beforeIncidentGitErr)

	os.Setenv(gitrepo.SimulateGitCommitFailureEnvvarName, "true")

	statusCode, _, _ := session.CreateIcon(moreDataIn[1].Name, moreDataIn[1].Iconfiles[0].Content)
	s.Equal(409, statusCode)

	afterIncidentSHA1, afterIncidentGitErr := s.TestGitRepo.GetStateID()
	s.NoError(afterIncidentGitErr)

	s.Equal(lastStableSHA1, afterIncidentSHA1)

	iconDescriptors, describeError := session.DescribeAllIcons()
	s.NoError(describeError)
	s.AssertResponseIconSetsEqual(dataOut, iconDescriptors)

	s.AssertEndState()
}
