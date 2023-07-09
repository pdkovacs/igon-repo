package git_tests

import (
	"testing"
	"time"

	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

// After adding support for distributed git-repo access, "remains consistent after failed <operation>"
// should mean that the operation can be successfully repeated after an initial failure.

func TestGitRepositoryTestSuite(t *testing.T) {
	for _, repo := range GitProvidersToTest() {
		suite.Run(t, &GitTestSuite{Repo: repo})
	}
}

func (s *GitTestSuite) TestAcceptsNewIconfileWhenEmpty() {
	var err error
	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]
	timeBeforeAdd := time.Now()
	err = s.Repo.AddIconfile(icon.Name, iconfile, icon.ModifiedBy)
	s.NoError(err)

	var sha1 string
	sha1, err = s.GetStateID()
	s.NoError(err)
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	s.AssertGitCleanStatus()
	s.AssertFileInRepo(icon.Name, iconfile, timeBeforeAdd)
}

func (s *GitTestSuite) TestAcceptsNewIconfileWhenNotEmpty() {
	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	errorWhenAddingFirstIconFile := s.Repo.AddIconfile(icon.Name, iconfile1, icon.ModifiedBy)
	s.NoError(errorWhenAddingFirstIconFile)

	firstSha1, errorWhenGettingFirstSha1 := s.GetStateID()
	s.NoError(errorWhenGettingFirstSha1)
	errorAddingSecondIconfile := s.Repo.AddIconfile(icon.Name, iconfile2, icon.ModifiedBy)
	s.NoError(errorAddingSecondIconfile)
	secondSha1, errorWhenGettingSecondSha1 := s.GetStateID()
	s.NoError(errorWhenGettingSecondSha1)
	s.NotEqual(firstSha1, secondSha1)
}

func (s *GitTestSuite) TestRemainsConsistentAfterUpdatingIconfileFails() {
}

func (s *GitTestSuite) TestRemainsConsistentAfterDeletingIconfileFails() {
}
