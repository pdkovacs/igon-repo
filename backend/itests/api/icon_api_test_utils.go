package api

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
)

type iconTestSuite struct {
	apiTestSuite
}

func (s *iconTestSuite) getCheckIconfile(session *apiTestSession, iconName string, iconfile domain.Iconfile) {
	actualIconfile, err := session.GetIconfile(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal(iconfile, actualIconfile)
}

func (s *iconTestSuite) assertGitCleanStatus() {
	s.testGitRepo.AssertGitCleanStatus(&s.Suite)
}

func (s *iconTestSuite) assertAllFilesInDBAreInGitAsWell() []string {
	checkedGitFiles := []string{}

	db := s.server.Repositories.DB
	git := s.testGitRepo

	allIconDesc, descAllErr := db.DescribeAllIcons()
	if descAllErr != nil {
		panic(descAllErr)
	}

	for _, iconDesc := range allIconDesc {
		for _, iconfileDesc := range iconDesc.Iconfiles {
			fileContentInDB, contentReadError := db.GetIconFile(iconDesc.Name, iconfileDesc.Format, iconfileDesc.Size)
			if contentReadError != nil {
				panic(contentReadError)
			}
			pathToFile := git.GetAbsolutePathToIconfile(iconDesc.Name, iconfileDesc)
			fileContentInGit, readGitFileErr := os.ReadFile(pathToFile)
			s.NoError(readGitFileErr)

			// TODO: fileContentInDB and fileContentInGit must equal
			s.True(bytes.Equal(fileContentInDB, fileContentInGit))

			checkedGitFiles = append(checkedGitFiles, s.testGitRepo.GetPathToIconfileInRepos(iconDesc.Name, iconfileDesc))
		}
	}

	return checkedGitFiles
}

func (s *iconTestSuite) createIconfilePaths(iconName string, iconfileDescriptor domain.IconfileDescriptor) web.IconPath {
	return web.CreateIconPath("/icon", iconName, iconfileDescriptor)
}

func (s *iconTestSuite) assertAllFilesInGitAreInDBAsWell(iconfilesWithPeerInDB []string) {
	iconfiles, err := s.testGitRepo.GetIconfiles()
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
			s.Fail(fmt.Sprintf("%s doesn't have a peer in DB", gitFile))
		}
	}
}

func (s *iconTestSuite) assertReposInSync() {
	checkedGitFiles := s.assertAllFilesInDBAreInGitAsWell()
	s.assertAllFilesInGitAreInDBAsWell(checkedGitFiles)
}

func (s *iconTestSuite) assertEndState() {
	s.assertGitCleanStatus()
	s.assertReposInSync()
}

func (s *iconTestSuite) assertResponseIconSetsEqual(expected []web.ResponseIcon, actual []web.ResponseIcon) {
	sortResponseIconSlice(expected)
	sortResponseIconSlice(actual)
	s.Equal(expected, actual)
}

func (s *iconTestSuite) assertResponseIconsEqual(expected web.ResponseIcon, actual web.ResponseIcon) {
	sortResponseIconPaths(expected)
	sortResponseIconPaths(actual)
	s.Equal(expected, actual)
}

func sortResponseIconSlice(slice []web.ResponseIcon) {
	sort.Slice(slice, func(i, j int) bool {
		return strings.Compare(slice[i].Name, slice[j].Name) < 0
	})
	for _, respIcon := range slice {
		sortResponseIconPaths(respIcon)
	}
}

func sortResponseIconPaths(respIcon web.ResponseIcon) {
	sort.Slice(respIcon.Paths, func(i, j int) bool {
		return strings.Compare(respIcon.Paths[i].Path, respIcon.Paths[j].Path) < 0
	})
	sort.Slice(respIcon.Tags, func(i, j int) bool {
		return strings.Compare(respIcon.Tags[i], respIcon.Tags[j]) < 0
	})
}
