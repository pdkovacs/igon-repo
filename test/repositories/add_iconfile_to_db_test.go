package repositories_tests

import (
	"errors"
	"testing"

	"igo-repo/internal/app/domain"
	itests_common "igo-repo/test/common"

	"github.com/stretchr/testify/suite"
)

type addIconfileToDBTestSuite struct {
	DBTestSuite
}

func TestAddIconfileToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconfileToDBTestSuite{})
}

func (s *addIconfileToDBTestSuite) TestErrorOnDuplicateIconfile() {
	var err error
	var icon = itests_common.TestData[0]
	var iconfile = icon.Iconfiles[0]

	err = s.dbRepo.CreateIcon(icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.dbRepo.AddIconfileToIcon(icon.Name, iconfile, icon.ModifiedBy, nil)
	s.True(errors.Is(err, domain.ErrIconfileAlreadyExists))
}

func (s *addIconfileToDBTestSuite) TestSecondIconfile() {
	var err error
	var icon = itests_common.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	err = s.dbRepo.CreateIcon(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.dbRepo.AddIconfileToIcon(icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *addIconfileToDBTestSuite) TestAddSecondIconfileBySecondUser() {
	var err error
	var icon = itests_common.TestData[0]
	var iconfile1 = icon.Iconfiles[0]
	var iconfile2 = icon.Iconfiles[1]

	var secondUser = "sedat"

	err = s.dbRepo.CreateIcon(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.dbRepo.AddIconfileToIcon(icon.Name, iconfile2, secondUser, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	clone := itests_common.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	s.equalIconAttributes(clone, iconDesc, nil)
}
