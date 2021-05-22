package db

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
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

	icon := testData[0]
	iconfile := icon.Iconfiles[0]

	err = repositories.CreateIcon(getPool(), icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)
	err = repositories.AddTag(getPool(), icon.Name, icon.Tags[0])
	s.NoError(err)

	err = repositories.DeleteIconfile(getPool(), icon.Name, iconfile, nil)
	s.NoError(err)

	_, err = repositories.DescribeIcon(getPool(), icon.Name)
	s.Error(domain.ErrIconNotFound, err)

	var rowCount int
	err = db.QueryRow("select count(*) as row_count from icon").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = db.QueryRow("select count(*) as row_count from icon_file").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = db.QueryRow("select count(*) as row_count from icon_to_tags").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
}

func (s *deleteIconfileFromDBTestSuite) TestDeleteNextToLastIconfile() {
	var err error

	icon := testData[0]
	iconfile1 := icon.Iconfiles[0]
	iconfile2 := icon.Iconfiles[1]

	err = repositories.CreateIcon(getPool(), icon.Name, iconfile1, icon.ModifiedBy, nil)
	s.NoError(err)
	err = repositories.AddTag(getPool(), icon.Name, icon.Tags[0])
	s.NoError(err)
	err = repositories.AddIconfileToIcon(getPool(), icon.Name, iconfile2, icon.ModifiedBy, nil)
	s.NoError(err)

	err = repositories.DeleteIconfile(getPool(), icon.Name, iconfile1, nil)
	s.NoError(err)

	var iconDesc domain.Icon
	iconDesc, err = repositories.DescribeIcon(getPool(), icon.Name)
	s.NoError(err)
	s.Equal(1, len(iconDesc.Iconfiles))
	s.equalIconAttributes(icon, iconDesc, nil)
}
