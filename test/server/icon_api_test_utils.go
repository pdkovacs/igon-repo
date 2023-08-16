package server

import (
	"fmt"
	"sort"
	"strings"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/httpadapter"
	"iconrepo/internal/repositories/blobstore/git"
	blobstore_tests "iconrepo/test/repositories/blobstore"
	"iconrepo/test/repositories/indexing"
)

type IconTestSuite struct {
	ApiTestSuite
}

func IconTestSuites(testSequenceId string) []IconTestSuite {
	all := []IconTestSuite{}
	for _, apiSuite := range apiTestSuites(
		testSequenceId, blobstore_tests.BlobstoreProvidersToTest(),
		indexing.IndexProvidersToTest(),
	) {
		all = append(all, IconTestSuite{ApiTestSuite: apiSuite})
	}
	return all
}

func (s *IconTestSuite) getCheckIconfile(session *apiTestSession, iconName string, iconfile domain.Iconfile) {
	actualIconfile, err := session.GetIconfile(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal(iconfile.Content, actualIconfile)
}

func (s *IconTestSuite) assertAllFilesIndexedAreInTheBlobstore() []string {
	checkedGitFiles := []string{}

	index := s.indexingController
	blob := s.TestBlobstoreController

	allIconDescIndexed, descAllErr := index.DescribeAllIcons(s.Ctx)
	if descAllErr != nil {
		s.FailNow("", "%v", descAllErr)
	}

	for _, iconDescIndexed := range allIconDescIndexed {
		for _, iconfileDesc := range iconDescIndexed.Iconfiles {
			fileContentInGit, readGitFileErr := blob.GetIconfile(iconDescIndexed.Name, iconfileDesc)
			s.NoError(readGitFileErr)
			s.Greater(len(fileContentInGit), 0)
			checkedGitFiles = append(checkedGitFiles, git.NewGitFilePaths("").GetPathToIconfileInRepo(iconDescIndexed.Name, iconfileDesc))
		}
	}

	return checkedGitFiles
}

func (s *IconTestSuite) createIconfilePaths(iconName string, iconfileDescriptor domain.IconfileDescriptor) httpadapter.IconPath {
	return httpadapter.CreateIconPath("/icon", iconName, iconfileDescriptor)
}

func (s *IconTestSuite) assertAllFilesInBlobstoreAreIndexed(iconfilesIndexed []string) {
	iconfiles, err := s.TestBlobstoreController.GetIconfiles()
	s.NoError(err)
	for _, gitFile := range iconfiles {
		found := false
		for _, dbFile := range iconfilesIndexed {
			if gitFile == dbFile {
				found = true
				break
			}
		}
		if !found {
			s.Fail(fmt.Sprintf("%s are not indexed (%#v)", gitFile, iconfilesIndexed))
		}
	}
}

func (s *IconTestSuite) assertReposInSync() {
	filesInBlobstore := s.assertAllFilesIndexedAreInTheBlobstore()
	s.assertAllFilesInBlobstoreAreIndexed(filesInBlobstore)
}

func (s *IconTestSuite) AssertEndState() {
	ok, err := s.TestBlobstoreController.CheckStatus()
	s.NoError(err)
	s.True(ok)
	s.assertReposInSync()
}

func (s *IconTestSuite) AssertResponseIconSetsEqual(expected []httpadapter.IconDTO, actual []httpadapter.IconDTO) {
	sortResponseIconSlice(expected)
	sortResponseIconSlice(actual)
	s.Equal(expected, actual)
}

func (s *IconTestSuite) assertResponseIconsEqual(expected httpadapter.IconDTO, actual httpadapter.IconDTO) {
	sortResponseIconPaths(expected)
	sortResponseIconPaths(actual)
	s.Equal(expected, actual)
}

func sortResponseIconSlice(slice []httpadapter.IconDTO) {
	sort.Slice(slice, func(i, j int) bool {
		return strings.Compare(slice[i].Name, slice[j].Name) < 0
	})
	for _, respIcon := range slice {
		sortResponseIconPaths(respIcon)
	}
}

func sortResponseIconPaths(respIcon httpadapter.IconDTO) {
	sort.Slice(respIcon.Paths, func(i, j int) bool {
		return strings.Compare(respIcon.Paths[i].Path, respIcon.Paths[j].Path) < 0
	})
	sort.Slice(respIcon.Tags, func(i, j int) bool {
		return strings.Compare(respIcon.Tags[i], respIcon.Tags[j]) < 0
	})
}
