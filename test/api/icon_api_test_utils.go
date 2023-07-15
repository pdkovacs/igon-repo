package api_tests

import (
	"fmt"
	"sort"
	"strings"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/httpadapter"
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/test/repositories/blobstore_tests/git_tests"
)

type IconTestSuite struct {
	ApiTestSuite
}

func IconTestSuites(testSequenceId string) []IconTestSuite {
	all := []IconTestSuite{}
	for _, apiSuite := range apiTestSuites(testSequenceId, git_tests.GitProvidersToTest()) {
		all = append(all, IconTestSuite{ApiTestSuite: apiSuite})
	}
	return all
}

func (s *IconTestSuite) getCheckIconfile(session *apiTestSession, iconName string, iconfile domain.Iconfile) {
	actualIconfile, err := session.GetIconfile(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal(iconfile.Content, actualIconfile)
}

func (s *IconTestSuite) assertAllFilesInDBAreInGitAsWell() []string {
	checkedGitFiles := []string{}

	index := s.indexRepo
	blob := s.TestBlobstore

	allIconDescInDb, descAllErr := index.DescribeAllIcons()
	if descAllErr != nil {
		panic(descAllErr)
	}

	for _, iconDescInDb := range allIconDescInDb {
		for _, iconfileDesc := range iconDescInDb.Iconfiles {
			fileContentInGit, readGitFileErr := blob.GetIconfile(iconDescInDb.Name, iconfileDesc)
			s.NoError(readGitFileErr)
			s.Greater(len(fileContentInGit), 0)
			checkedGitFiles = append(checkedGitFiles, git.NewGitFilePaths("").GetPathToIconfileInRepo(iconDescInDb.Name, iconfileDesc))
		}
	}

	return checkedGitFiles
}

func (s *IconTestSuite) createIconfilePaths(iconName string, iconfileDescriptor domain.IconfileDescriptor) httpadapter.IconPath {
	return httpadapter.CreateIconPath("/icon", iconName, iconfileDescriptor)
}

func (s *IconTestSuite) assertAllFilesInGitAreInDBAsWell(iconfilesWithPeerInDB []string) {
	iconfiles, err := s.TestBlobstore.GetIconfiles()
	s.NoError(err)
	for _, gitFile := range iconfiles {
		found := false
		for _, dbFile := range iconfilesWithPeerInDB {
			if gitFile == dbFile {
				found = true
				break
			}
		}
		if !found {
			s.Fail(fmt.Sprintf("%s doesn't have a peer in DB (%#v)", gitFile, iconfilesWithPeerInDB))
		}
	}
}

func (s *IconTestSuite) assertReposInSync() {
	checkedGitFiles := s.assertAllFilesInDBAreInGitAsWell()
	s.assertAllFilesInGitAreInDBAsWell(checkedGitFiles)
}

func (s *IconTestSuite) AssertEndState() {
	ok, err := s.TestBlobstore.CheckStatus()
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
