package repositories

import (
	"testing"

	"github.com/pdkovacs/igo-repo/internal/domain"
	itests_common "github.com/pdkovacs/igo-repo/test/common"
	"github.com/stretchr/testify/suite"
)

type deleteIconfileFromDBTestSuite struct {
	DBTestSuite
}

func TestDeleteIconfileFromDBTestSuite(t *testing.T) {
	suite.Run(t, &deleteIconfileFromDBTestSuite{})
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteTheOnlyIconfile() {
	var err error

	icon := itests_common.TestData[0]
	iconfile := icon.Iconfiles[0]

	err = s.dbRepo.CreateIcon(icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.AddTag(icon.Name, icon.ModifiedBy, icon.Tags[0])
	s.NoError(err)

	err = s.dbRepo.DeleteIconfile(icon.Name, iconfile.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	_, err = s.dbRepo.DescribeIcon(icon.Name)
	s.Error(domain.ErrIconNotFound, err)

	var rowCount int
	err = s.dbRepo.ConnectionPool.QueryRow("select count(*) as row_count from icon").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = s.dbRepo.ConnectionPool.QueryRow("select count(*) as row_count from icon_file").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = s.dbRepo.ConnectionPool.QueryRow("select count(*) as row_count from icon_to_tags").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfile() {
	var err error

	icon := itests_common.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.dbRepo.CreateIcon(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.dbRepo.AddIconfileToIcon(icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.dbRepo.DeleteIconfile(icon.Name, iconfile1.IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfileBySecondUser() {
	var err error

	const secondUser = "second-user"

	icon := itests_common.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.dbRepo.CreateIcon(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.dbRepo.AddIconfileToIcon(icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.dbRepo.DeleteIconfile(icon.Name, iconfile1.IconfileDescriptor, secondUser, nil)
	s.NoError(err)

	clone := itests_common.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	var iconDesc domain.IconDescriptor
	iconDesc, err = s.dbRepo.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(clone, iconDesc, nil)
}
