package repositories

import (
	"os"
	"testing"

	"igo-repo/internal/repositories"
	itests_common "igo-repo/test/common"

	"github.com/stretchr/testify/suite"
)

func TestgitRepositoryTestSuite(t *testing.T) {
	suite.Run(t, &GitTestSuite{})
}

func (s *GitTestSuite) TestAcceptsNewIconfile() {
	var err error
	icon := itests_common.TestData[0]
	iconfile := icon.Iconfiles[0]
	err = s.repo.AddIconfile(icon.Name, iconfile, icon.ModifiedBy)
	s.NoError(err)

	var sha1 string
	sha1, err = s.getCurrentCommit()
	s.NoError(err)
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	s.assertGitCleanStatus()
	s.assertFileInRepo(icon.Name, iconfile)
}

func (s *GitTestSuite) TestAcceptsMoreIconfiles() {
	icon := itests_common.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	errorWhenAddingFirstIconFile := s.repo.AddIconfile(icon.Name, iconfile1, icon.ModifiedBy)
	s.NoError(errorWhenAddingFirstIconFile)

	firstSha1, errorWhenGettingFirstSha1 := s.getCurrentCommit()
	s.NoError(errorWhenGettingFirstSha1)
	errorAddingSecondIconfile := s.repo.AddIconfile(icon.Name, iconfile2, icon.ModifiedBy)
	s.NoError(errorAddingSecondIconfile)
	secondSha1, errorWhenGettingSecondSha1 := s.getCurrentCommit()
	s.NoError(errorWhenGettingSecondSha1)
	s.NotEqual(firstSha1, secondSha1)
}

func (s *GitTestSuite) TestRemainsConsistentAfterAddingIconfileFails() {
	icon := itests_common.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	errorWhenAddingFirstIconFile := s.repo.AddIconfile(icon.Name, iconfile1, icon.ModifiedBy)
	s.NoError(errorWhenAddingFirstIconFile)

	os.Setenv(repositories.IntrusiveGitTestEnvvarName, "true")

	lastGoodSha1, errorWhenGettingLastGoodSha1 := s.getCurrentCommit()
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(lastGoodSha1))
	s.NoError(errorWhenGettingLastGoodSha1)
	errorAddingSecondIconfile := s.repo.AddIconfile(icon.Name, iconfile2, icon.ModifiedBy)
	s.Error(errorAddingSecondIconfile)
	postSha1, errorWhenGettingPostSha1 := s.getCurrentCommit()
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(postSha1))
	s.NoError(errorWhenGettingPostSha1)
	s.Equal(lastGoodSha1, postSha1)
	s.assertGitCleanStatus()
	s.assertFileInRepo(icon.Name, iconfile1)
	s.assertFileNotInRepo(icon.Name, iconfile2)
}
