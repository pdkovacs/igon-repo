package repositories

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/itests"
	"github.com/stretchr/testify/suite"
)

type deleteIconfileFromDBTestSuite struct {
	dbTestSuite
}

func TestDeleteIconfileFromDBTestSuite(t *testing.T) {
	suite.Run(t, &deleteIconfileFromDBTestSuite{})
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteTheOnlyIconfile() {
	var err error

	icon := itests.TestData[0]
	iconfile := icon.Iconfiles[0]

	err = s.CreateIcon(icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.AddTag(icon.Name, icon.ModifiedBy, icon.Tags[0])
	s.NoError(err)

	err = s.DeleteIconfile(icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)

	_, err = s.DescribeIcon(icon.Name)
	s.Error(domain.ErrIconNotFound, err)

	var rowCount int
	err = s.ConnectionPool.QueryRow("select count(*) as row_count from icon").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = s.ConnectionPool.QueryRow("select count(*) as row_count from icon_file").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = s.ConnectionPool.QueryRow("select count(*) as row_count from icon_to_tags").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfile() {
	var err error

	icon := itests.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.CreateIcon(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.AddIconfileToIcon(icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.DeleteIconfile(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)

	var iconDesc domain.Icon
	iconDesc, err = s.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(icon, iconDesc, nil)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfileBySecondUser() {
	var err error

	const secondUser = "second-user"

	icon := itests.TestData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = s.CreateIcon(icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)
	err = s.AddIconfileToIcon(icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	err = s.DeleteIconfile(icon.Name, iconfile1, secondUser, nil)
	s.NoError(err)

	clone := itests.CloneIcon(icon)
	clone.ModifiedBy = secondUser
	var iconDesc domain.Icon
	iconDesc, err = s.DescribeIcon(icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(clone, iconDesc, nil)
}
