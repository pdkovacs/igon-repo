package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addIconToDBTestSuite struct {
	dbTestSuite
}

func TestAddIconToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconToDBTestSuite{})
}

func (s *addIconToDBTestSuite) TestAddFirstIcon() {
	var icon = testData[0]
	fmt.Printf("Hello, First Icon %v\n", icon.Name)
	err := repositories.CreateIcon(getPool(), icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	var iconDesc domain.Icon
	iconDesc, err = repositories.DescribeIcon(getPool(), icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
	s.getIconfileChecked(icon.Name, icon.Iconfiles[0])
}

func (s *addIconToDBTestSuite) TestAddASecondIcon() {
	var err error
	var icon1 = testData[0]
	var icon2 = testData[1]
	err = repositories.CreateIcon(getPool(), icon1.Name, icon1.Iconfiles[0], icon1.ModifiedBy, nil)
	s.NoError(err)
	err = repositories.CreateIcon(getPool(), icon2.Name, icon2.Iconfiles[1], icon2.ModifiedBy, nil)
	s.NoError(err)
	var count int
	count, err = getIconCount()
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

	var icon1 = testData[0]
	var icon2 = testData[1]
	err = repositories.CreateIcon(getPool(), icon1.Name, icon1.Iconfiles[0], icon1.ModifiedBy, nil)
	s.NoError(err)
	err = repositories.CreateIcon(getPool(), icon2.Name, icon2.Iconfiles[1], icon2.ModifiedBy, createSideEffect)
	s.True(errors.Is(err, sideEffectTestError))

	count, err = getIconCount()
	s.NoError(err)
	s.Equal(1, count)
	s.getIconfileChecked(icon1.Name, icon1.Iconfiles[0])
	_, err = s.getIconfile(icon2.Name, icon2.Iconfiles[1])
	s.True(errors.Is(err, domain.ErrIconNotFound))
}
