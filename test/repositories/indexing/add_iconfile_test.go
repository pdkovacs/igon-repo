package indexing

import (
	"errors"
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type addIconfileToDBTestSuite struct {
	IndexingTestSuite
}

func TestAddIconfileToDBTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &addIconfileToDBTestSuite{testSuite})
	}
}

func (s *addIconfileToDBTestSuite) TestErrorOnDuplicateIconfile() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile = icon.Iconfiles[0]

	err = s.testRepoController.CreateIcon(icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddIconfileToIcon(icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.True(errors.Is(err, domain.ErrIconfileAlreadyExists))
}

func (s *addIconfileToDBTestSuite) TestSecondIconfile() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	err = s.testRepoController.CreateIcon(icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddIconfileToIcon(icon.Name, iconfile2.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *addIconfileToDBTestSuite) TestAddSecondIconfileBySecondUser() {
	var err error
	var icon = test_commons.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	var secondUser = "sedat"

	err = s.testRepoController.CreateIcon(icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.AddIconfileToIcon(icon.Name, iconfile2.IconfileDescriptor, secondUser, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(icon.Name)
	s.NoError(err)
	clone := test_commons.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	s.equalIconAttributes(clone, iconDesc, nil)
}
