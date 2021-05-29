package repositories

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/itests"
	"github.com/stretchr/testify/suite"
)

type deleteIconFromDBTestSuite struct {
	dbTestSuite
}

func TestDeleteIconFromDBTestSuite(t *testing.T) {
	suite.Run(t, &deleteIconFromDBTestSuite{})
}

func (s *deleteIconFromDBTestSuite) TestDeleteAllAssociatedEntries() {
	var err error

	icon := itests.TestData[0]

	err = s.CreateIcon(icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.DeleteIcon(icon.Name, icon.ModifiedBy, nil)
	s.NoError(err)

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
