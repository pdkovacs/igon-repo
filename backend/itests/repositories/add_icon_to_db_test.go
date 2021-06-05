package repositories

import (
	"errors"
	"fmt"
	"testing"

	itest_common "github.com/pdkovacs/igo-repo/backend/itests/common"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/stretchr/testify/suite"
)

type addIconToDBTestSuite struct {
	DBTestSuite
}

func TestAddIconToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconToDBTestSuite{})
}

func (s *addIconToDBTestSuite) TestAddFirstIcon() {
	var icon = itest_common.TestData[0]
	fmt.Printf("Hello, First Icon %v\n", icon.Name)
	err := s.dbRepo.CreateIcon(icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	var iconDesc domain.Icon
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
	s.getIconfileChecked(icon.Name, icon.Iconfiles[0])
}

func (s *addIconToDBTestSuite) TestAddASecondIcon() {
	var err error
	var icon1 = itest_common.TestData[0]
	var icon2 = itest_common.TestData[1]
	err = s.dbRepo.CreateIcon(icon1.Name, icon1.Iconfiles[0], icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.CreateIcon(icon2.Name, icon2.Iconfiles[1], icon2.ModifiedBy, nil)
	s.NoError(err)
	var count int
	count, err = s.getIconCount()
	s.NoError(err)
	s.Equal(2, count)
	s.getIconfileChecked(icon1.Name, icon1.Iconfiles[0])
	s.getIconfileChecked(icon2.Name, icon2.Iconfiles[1])
}

// should rollback to last consistent state, in case an error occurs in sideEffect
func (s *addIconToDBTestSuite) TestRollbackOnErrorInSideEffect() {
	var count int
	var err error

	var sideEffectTestError = errors.New("some error occurred in side-effect")
	var createSideEffect = func() error {
		return sideEffectTestError
	}

	var icon1 = itest_common.TestData[0]
	var icon2 = itest_common.TestData[1]
	err = s.dbRepo.CreateIcon(icon1.Name, icon1.Iconfiles[0], icon1.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.CreateIcon(icon2.Name, icon2.Iconfiles[1], icon2.ModifiedBy, createSideEffect)
	s.True(errors.Is(err, sideEffectTestError))

	count, err = s.getIconCount()
	s.NoError(err)
	s.Equal(1, count)
	s.getIconfileChecked(icon1.Name, icon1.Iconfiles[0])
	_, err = s.getIconfile(icon2.Name, icon2.Iconfiles[1])
	s.Error(err)
	s.True(errors.Is(err, domain.ErrIconfileNotFound))
	if !errors.Is(err, domain.ErrIconfileNotFound) {
		s.Equal("blabla", err)
	}
}
