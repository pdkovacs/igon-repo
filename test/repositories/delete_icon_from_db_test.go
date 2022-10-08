package repositories_tests

import (
	"testing"

	itests_common "igo-repo/test/common"

	"github.com/stretchr/testify/suite"
)

type deleteIconFromDBTestSuite struct {
	DBTestSuite
}

func TestDeleteIconFromDBTestSuite(t *testing.T) {
	suite.Run(t, &deleteIconFromDBTestSuite{})
}

func (s *deleteIconFromDBTestSuite) TestDeleteAllAssociatedEntries() {
	var err error

	icon := itests_common.TestData[0]

	err = s.dbRepo.CreateIcon(icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.dbRepo.DeleteIcon(icon.Name, icon.ModifiedBy, nil)
	s.NoError(err)

	var rowCount int
	err = s.dbRepo.Conn.Pool.QueryRow("select count(*) as row_count from icon").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = s.dbRepo.Conn.Pool.QueryRow("select count(*) as row_count from icon_file").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
	err = s.dbRepo.Conn.Pool.QueryRow("select count(*) as row_count from icon_to_tags").Scan(&rowCount)
	s.NoError(err)
	s.Equal(0, rowCount)
}
