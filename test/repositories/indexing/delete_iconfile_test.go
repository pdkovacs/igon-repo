package indexing

import (
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type deleteIconfileFromDBTestSuite struct {
	IndexingTestSuite
}

func TestDeleteIconfileFromDBTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &deleteIconfileFromDBTestSuite{testSuite})
	}
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteTheOnlyIconfile() {
	var err error

	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]

	err = s.testRepoController.CreateIcon(icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(icon.Name, icon.ModifiedBy, icon.Tags[0])
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	_, err = s.testRepoController.DescribeIcon(icon.Name)
	s.Error(domain.ErrIconNotFound, err)

	var rowCount int
	rowCount, err = s.testRepoController.GetIconCount()
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.testRepoController.GetIconFileCount()
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.testRepoController.GetTagRelationCount()
	s.NoError(err)
	s.Equal(0, rowCount)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfile() {
	var err error

	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.testRepoController.CreateIcon(icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.testRepoController.AddIconfileToIcon(icon.Name, iconfile2.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfileBySecondUser() {
	var err error

	const secondUser = "second-user"

	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.testRepoController.CreateIcon(icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.testRepoController.AddIconfileToIcon(icon.Name, iconfile2.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(icon.Name, iconfile1.IconfileDescriptor, secondUser, nil)
	s.NoError(err)

	clone := test_commons.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(clone, iconDesc, nil)
}
