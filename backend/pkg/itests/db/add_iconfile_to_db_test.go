package db

import (
	"errors"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addIconfileToDBTestSuite struct {
	dbTestSuite
}

func TestAddIconfileToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconfileToDBTestSuite{})
}

func (s *addIconfileToDBTestSuite) TestErrorOnDuplicateIconfile() {
	var err error
	var icon = testData[0]
	var iconfile = icon.Iconfiles[0]

	err = repositories.CreateIcon(getPool(), icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)

	err = repositories.AddIconfileToIcon(getPool(), icon.Name, iconfile, icon.ModifiedBy, nil)
	s.True(errors.Is(err, domain.ErrIconfileAlreadyExists))
}

func (s *addIconfileToDBTestSuite) TestSecondIconfile() {
	var err error
	var icon = testData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	err = repositories.CreateIcon(getPool(), icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)

	err = repositories.AddIconfileToIcon(getPool(), icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.Icon
	iconDesc, err = repositories.DescribeIcon(getPool(), icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *addIconfileToDBTestSuite) TestAddSecondIconfileBySecondUser() {
	var err error
	var icon = testData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	var secondUser = "sedat"

	err = repositories.CreateIcon(getPool(), icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)

	err = repositories.AddIconfileToIcon(getPool(), icon.Name, iconfile2, secondUser, nil)
	s.NoError(err)

	var iconDesc domain.Icon
	iconDesc, err = repositories.DescribeIcon(getPool(), icon.Name)
	s.NoError(err)
	clone := cloneIcon(icon)
	clone.ModifiedBy = secondUser
	s.equalIconAttributes(clone, iconDesc, nil)
}
