package repositories

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/itests"
	"github.com/stretchr/testify/suite"
)

func TestGitRepositoryTestSuite(t *testing.T) {
	suite.Run(t, &gitTestSuite{})
}

func (s *gitTestSuite) TestAcceptsNewIconfile() {
	var err error
	icon := itests.TestData[0]
	iconfile := icon.Iconfiles[0]
	err = s.Repo.AddIconfile(icon.Name, iconfile, icon.ModifiedBy)
	s.NoError(err)

	var sha1 string
	sha1, err = s.getCurrentCommit()
	s.NoError(err)
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	s.assertGitCleanStatus()
	s.assertFileInRepo(icon.Name, iconfile)
}

func (s *gitTestSuite) TestRemainsConsistentAfterAddingIconfileFails() {

}
