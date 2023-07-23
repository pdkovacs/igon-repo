package indexing

import (
	"errors"
	"fmt"
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type addIconToIndexTestSuite struct {
	IndexingTestSuite
}

func TestAddIconToIndexTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &addIconToIndexTestSuite{testSuite})
	}
}

func (s *addIconToIndexTestSuite) TestAddFirstIconToIndex() {
	var icon = test_commons.TestData[0]
	fmt.Printf("Hello, First Icon %v\n", icon.Name)
	err := s.testRepoController.CreateIcon(s.ctx, icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *addIconToIndexTestSuite) TestAddASecondIconToIndex() {
	var err error
	var icon1 = test_commons.TestData[0]
	var icon2 = test_commons.TestData[1]
	err = s.testRepoController.CreateIcon(s.ctx, icon1.Name, icon1.Iconfiles[0].IconfileDescriptor, icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.CreateIcon(s.ctx, icon2.Name, icon2.Iconfiles[1].IconfileDescriptor, icon2.ModifiedBy, nil)
	s.NoError(err)
	var count int
	count, err = s.getIconCount(s.ctx)
	s.NoError(err)
	s.Equal(2, count)
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon2.Name)
	s.NoError(err)
	s.equalIconAttributes(icon2, iconDesc, nil)
}

// should rollback to last consistent state, in case an error occurs in sideEffect
func (s *addIconToIndexTestSuite) TestRollbackOnFailedSideEffect() {
	var count int
	var err error

	var createSideEffect = func() error {
		return errSideEffectTest
	}

	var icon1 = test_commons.TestData[0]
	var icon2 = test_commons.TestData[1]
	err = s.testRepoController.CreateIcon(s.ctx, icon1.Name, icon1.Iconfiles[0].IconfileDescriptor, icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.CreateIcon(s.ctx, icon2.Name, icon2.Iconfiles[1].IconfileDescriptor, icon2.ModifiedBy, createSideEffect)
	s.Error(err)
	s.ErrorIs(err, errSideEffectTest)

	count, err = s.getIconCount(s.ctx)
	s.NoError(err)
	s.Equal(1, count)

	var iconDesc domain.IconDescriptor

	iconDesc, err = s.testRepoController.DescribeIcon(s.ctx, icon1.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(icon1, iconDesc, nil)

	_, err = s.testRepoController.DescribeIcon(s.ctx, icon2.Name)
	s.Error(err)
	s.ErrorIs(err, domain.ErrIconNotFound)

	if !errors.Is(err, domain.ErrIconNotFound) {
		s.Equal("blabla", err)
	}
}
