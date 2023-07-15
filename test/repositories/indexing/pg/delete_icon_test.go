package pg

import (
	"iconrepo/test/test_commons"
	"testing"

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

	icon := test_commons.TestData[0]

	err = s.dbRepo.CreateIcon(icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.dbRepo.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.dbRepo.DeleteIcon(icon.Name, icon.ModifiedBy, nil)
	s.NoError(err)

	var rowCount int
	rowCount, err = s.dbRepo.GetIconCount()
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.dbRepo.GetIconFileCount()
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.dbRepo.GetTagRelationCount()
	s.NoError(err)
	s.Equal(0, rowCount)
}
