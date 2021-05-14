package db

import (
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addTagTestSuite struct {
	dbTestSuite
}

func TestAddTagTestSuite(t *testing.T) {
	suite.Run(t, &addTagTestSuite{})
}

func (s *addTagTestSuite) TestCreateAssociateNonExistingTag() {
	var tags []string
	var err error

	var icon = testData[0]
	const tag = "used-in-marvinjs"

	err = repositories.CreateIcon(getPool(), icon.Name, icon.Iconfiles[0], icon.ModifiedBy, nil)
	s.NoError(err)
	tags, err = repositories.GetExistingTags(getPool())
	s.NoError(err)
	s.Empty(tags)
}
