package pg

import (
	"errors"
	"fmt"
	"testing"

	"iconrepo/internal/app/domain"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type addIconToDBTestSuite struct {
	DBTestSuite
}

func TestAddIconToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconToDBTestSuite{})
}

func (s *addIconToDBTestSuite) TestAddFirstIcon() {
	var icon = test_commons.TestData[0]
	fmt.Printf("Hello, First Icon %v\n", icon.Name)
	err := s.dbRepo.CreateIcon(icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *addIconToDBTestSuite) TestAddASecondIcon() {
	var err error
	var icon1 = test_commons.TestData[0]
	var icon2 = test_commons.TestData[1]
	err = s.dbRepo.CreateIcon(icon1.Name, icon1.Iconfiles[0].IconfileDescriptor, icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.CreateIcon(icon2.Name, icon2.Iconfiles[1].IconfileDescriptor, icon2.ModifiedBy, nil)
	s.NoError(err)
	var count int
	count, err = s.getIconCount()
	s.NoError(err)
	s.Equal(2, count)
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon2.Name)
	s.NoError(err)
	s.equalIconAttributes(icon2, iconDesc, nil)
}

// should rollback to last consistent state, in case an error occurs in sideEffect
func (s *addIconToDBTestSuite) TestRollbackOnErrorInSideEffect() {
	var count int
	var err error

	var sideEffectTestError = errors.New("some error occurred in side-effect")
	var createSideEffect = func() error {
		return sideEffectTestError
	}

	var icon1 = test_commons.TestData[0]
	var icon2 = test_commons.TestData[1]
	err = s.dbRepo.CreateIcon(icon1.Name, icon1.Iconfiles[0].IconfileDescriptor, icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.CreateIcon(icon2.Name, icon2.Iconfiles[1].IconfileDescriptor, icon2.ModifiedBy, createSideEffect)
	s.True(errors.Is(err, sideEffectTestError))

	count, err = s.getIconCount()
	s.NoError(err)
	s.Equal(1, count)

	var iconDesc domain.IconDescriptor

	iconDesc, err = s.dbRepo.DescribeIcon(icon1.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(icon1, iconDesc, nil)

	_, err = s.dbRepo.DescribeIcon(icon2.Name)
	s.Error(err)
	s.True(errors.Is(err, domain.ErrIconNotFound))
	if !errors.Is(err, domain.ErrIconNotFound) {
		s.Equal("blabla", err)
	}
}
