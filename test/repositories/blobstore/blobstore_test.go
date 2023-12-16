package blobstore

import (
	"context"
	"testing"
	"time"

	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

// After adding support for distributed git-repo access, "remains consistent after failed <operation>"
// should mean that the operation can be successfully repeated after an initial failure.

func TestGitRepositoryTestSuite(t *testing.T) {
	for _, repoController := range BlobstoreProvidersToTest() {
		suite.Run(t, &BlobstoreTestSuite{RepoController: repoController, TestSequenceId: "blobstore", Ctx: context.Background()})
	}
}

func (s *BlobstoreTestSuite) TestAcceptsNewIconfileWhenEmpty() {
	var err error
	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]
	timeBeforeAdd := time.Now()
	err = s.RepoController.AddIconfile(s.Ctx, icon.Name, iconfile, icon.ModifiedBy)
	s.NoError(err)

	var sha1 string
	sha1, err = s.GetStateID()
	s.NoError(err)
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	s.AssertBlobstoreCleanStatus()
	s.AssertFileInBlobstore(icon.Name, iconfile, timeBeforeAdd)
}

func (s *BlobstoreTestSuite) TestAcceptsNewIconfileWhenNotEmpty() {
	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	errorWhenAddingFirstIconFile := s.RepoController.AddIconfile(s.Ctx, icon.Name, iconfile1, icon.ModifiedBy)
	s.NoError(errorWhenAddingFirstIconFile)

	firstSha1, errorWhenGettingFirstSha1 := s.GetStateID()
	s.NoError(errorWhenGettingFirstSha1)
	errorAddingSecondIconfile := s.RepoController.AddIconfile(s.Ctx, icon.Name, iconfile2, icon.ModifiedBy)
	s.NoError(errorAddingSecondIconfile)
	secondSha1, errorWhenGettingSecondSha1 := s.GetStateID()
	s.NoError(errorWhenGettingSecondSha1)
	s.NotEqual(firstSha1, secondSha1)
}

func (s *BlobstoreTestSuite) TestRemainsConsistentAfterUpdatingIconfileFails() {
}

func (s *BlobstoreTestSuite) TestRemainsConsistentAfterDeletingIconfileFails() {
}
