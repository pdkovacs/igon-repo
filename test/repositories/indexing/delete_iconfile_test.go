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

func TestDeleteIconfileFromIndexTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &deleteIconfileFromDBTestSuite{testSuite})
	}
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteTheOnlyIconfile() {
	var err error

	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(s.ctx, icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(s.ctx, icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	_, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.Error(domain.ErrIconNotFound, err)

	var rowCount int
	rowCount, err = s.testRepoController.GetIconCount(s.ctx)
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.testRepoController.GetIconFileCount(s.ctx)
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.testRepoController.GetTagRelationCount(s.ctx)
	s.NoError(err)
	s.Equal(0, rowCount)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfile() {
	var err error

	icon := test_commons.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(s.ctx, icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.testRepoController.AddIconfileToIcon(s.ctx, icon.Name, iconfile2.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(s.ctx, icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
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

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(s.ctx, icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.testRepoController.AddIconfileToIcon(s.ctx, icon.Name, iconfile2.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(s.ctx, icon.Name, iconfile1.IconfileDescriptor, secondUser, nil)
	s.NoError(err)

	clone := test_commons.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(clone, iconDesc, nil)
}

func (s *deleteIconfileFromDBTestSuite) TestRollbackOnFailedSideEffect() {
	var err error

	icon := test_commons.TestData[0]
	iconfile := icon.Iconfiles[0]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(s.ctx, icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.testRepoController.DeleteIconfile(s.ctx, icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, func() error {
		return errSideEffectTest
	})
	s.Error(err)
	s.ErrorIs(err, errSideEffectTest)

	iconDescArr, describeErr := s.testRepoController.DescribeAllIcons(s.ctx)
	s.NoError(describeErr)
	s.Equal(1, len(iconDescArr))
	s.equalIconAttributes(icon, iconDescArr[0], nil)
}
