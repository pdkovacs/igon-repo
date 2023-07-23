package indexing

import (
	"iconrepo/test/test_commons"
	"testing"

	"github.com/stretchr/testify/suite"
)

type deleteIconFromIndexTestSuite struct {
	IndexingTestSuite
}

func TestDeleteIconFromIndexTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &deleteIconFromIndexTestSuite{testSuite})
	}
}

func (s *deleteIconFromIndexTestSuite) TestDeleteAllAssociatedEntries() {
	var err error

	icon := test_commons.TestData[0]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(s.ctx, icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.testRepoController.DeleteIcon(s.ctx, icon.Name, icon.ModifiedBy, nil)
	s.NoError(err)

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

func (s *deleteIconFromIndexTestSuite) TestRollbackOnFailedSideEffect() {
	var err error

	icon := test_commons.TestData[0]

	err = s.testRepoController.CreateIcon(s.ctx, icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(s.ctx, icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.testRepoController.DeleteIcon(s.ctx, icon.Name, icon.ModifiedBy, func() error {
		return errSideEffectTest
	})
	s.Error(err)
	s.ErrorIs(err, errSideEffectTest)

	iconDescArr, describeErr := s.testRepoController.DescribeAllIcons(s.ctx)
	s.NoError(describeErr)
	s.Equal(1, len(iconDescArr))
	s.equalIconAttributes(icon, iconDescArr[0], nil)
}
