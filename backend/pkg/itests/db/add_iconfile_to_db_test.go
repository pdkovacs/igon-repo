package db

import (
	"errors"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addIconfileToDBTestSuite struct {
	suite.Suite
}

func TestAddIconfileToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconfileToDBTestSuite{})
}

func (s *addIconfileToDBTestSuite) SetupSuite() {
	manageTestResourcesBeforeAll()
}

func (s *addIconfileToDBTestSuite) TearDownSuite() {
	manageTestResourcesAfterAll()
}

func (s *addIconfileToDBTestSuite) BeforeTest(suiteName, testName string) {
	manageTestResourcesBeforeEach()
}

func (s *addIconfileToDBTestSuite) AfterTest(suiteName, testName string) {
}

func (s *addIconfileToDBTestSuite) TestErrorOnDuplicateIconfile() {
	var err error
	const user = "Zazie"

	var iconfile1 = createTestIconfile("metro-icon", "french", "great")
	err = repositories.CreateIcon(getPool(), iconfile1, user, nil)
	s.NoError(err)

	var iconfile2 = createTestIconfile("metro-icon", "french", "great")
	err = repositories.AddIconfileToIcon(getPool(), iconfile2, user, nil)
	s.True(errors.Is(err, domain.ErrIconfileAlreadyExists))
}
