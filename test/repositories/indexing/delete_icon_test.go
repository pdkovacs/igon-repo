package indexing

import (
	"iconrepo/test/test_commons"
	"testing"

	"github.com/stretchr/testify/suite"
)

type deleteIconFromDBTestSuite struct {
	IndexingTestSuite
}

func TestDeleteIconFromDBTestSuite(t *testing.T) {
	for _, testSuite := range indexingTestSuites() {
		suite.Run(t, &deleteIconFromDBTestSuite{testSuite})
	}
}

func (s *deleteIconFromDBTestSuite) TestDeleteAllAssociatedEntries() {
	var err error

	icon := test_commons.TestData[0]

	err = s.testRepoController.CreateIcon(icon.Name, icon.Iconfiles[0].IconfileDescriptor, icon.ModifiedBy, nil)
	s.NoError(err)
	err = s.testRepoController.AddTag(icon.Name, icon.Tags[0], icon.ModifiedBy)
	s.NoError(err)

	err = s.testRepoController.DeleteIcon(icon.Name, icon.ModifiedBy)
	s.NoError(err)

	var rowCount int
	rowCount, err = s.testRepoController.GetIconCount()
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.testRepoController.GetIconFileCount()
	s.NoError(err)
	s.Equal(0, rowCount)
	rowCount, err = s.testRepoController.GetTagRelationCount()
	s.NoError(err)
	s.Equal(0, rowCount)
}
