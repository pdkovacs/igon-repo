package db

import (
	"errors"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addIconfileToDBTestSuite struct {
	dbTestSuite
}

func TestAddIconfileToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconfileToDBTestSuite{})
}

func (s *addIconfileToDBTestSuite) TestErrorOnDuplicateIconfile() {
	var err error
	var icon = testData[0]
	var iconfile = icon.Iconfiles[0]

	err = repositories.CreateIcon(getPool(), icon.Name, iconfile, icon.ModifiedBy, nil)
	s.NoError(err)

	err = repositories.AddIconfileToIcon(getPool(), icon.Name, iconfile, icon.ModifiedBy, nil)
	s.True(errors.Is(err, domain.ErrIconfileAlreadyExists))
}
