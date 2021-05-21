package db

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
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

	icon := testData[0]

	err = repositories.CreateIcon(getPool(), icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	err = repositories.AddTag(getPool(), icon.Name, icon.Tags[0])
	s.NoError(err)

	err = repositories.DeleteIcon(getPool(), icon.Name, nil)
	s.NoError(err)

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
