package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/stretchr/testify/suite"
)

type addIconToDBTestSuite struct {
	suite.Suite
}

func TestAddIconToDBTestSuite(t *testing.T) {
	suite.Run(t, &addIconToDBTestSuite{})
}

func (s *addIconToDBTestSuite) BeforeTest(suiteName, testName string) {
	manageTestResourcesBeforeEach()
}

func (s *addIconToDBTestSuite) AfterTest(suiteName, testName string) {
}

func (s *addIconToDBTestSuite) SetupSuite() {
	manageTestResourcesBeforeAll()
}

func (s *addIconToDBTestSuite) TearDownSuite() {
	manageTestResourcesAfterAll()
}

func (s *addIconToDBTestSuite) getIconfile(iconfile domain.Iconfile) ([]byte, error) {
	return repositories.GetIconFile(getPool(), iconfile.Name, iconfile.Format, iconfile.Size)
}

func (s *addIconToDBTestSuite) getIconfileChecked(iconfile domain.Iconfile) {
	content, err := s.getIconfile(iconfile)
	s.NoError(err)
	s.Equal(iconfile.Content, content)
}

func (s *addIconToDBTestSuite) TestAddFirstIcon() {
	const user = "zazie"
	var iconfile = createTestIconfile("metro-icon", "french", "great")
	fmt.Printf("Hello, First Icon %v\n", iconfile.Name)
	err := repositories.CreateIcon(getPool(), iconfile, user, nil)
	s.NoError(err)
	s.getIconfileChecked(iconfile)
}

func (s *addIconToDBTestSuite) TestAddASecondIcon() {
	var err error
	const user = "zazie"
	var iconfile1 = createTestIconfile("metro-icon", "french", "great")
	var iconfile2 = createTestIconfile("animal-icon", "french", "huge")
	err = repositories.CreateIcon(getPool(), iconfile1, user, nil)
	s.NoError(err)
	err = repositories.CreateIcon(getPool(), iconfile2, user, nil)
	s.NoError(err)
	var count int
	count, err = getIconCount()
	s.NoError(err)
	s.Equal(2, count)
	s.getIconfileChecked(iconfile1)
	s.getIconfileChecked(iconfile2)
}

// should rollback to last consistent state, in case an error occurs in sideEffect
func (s *addIconToDBTestSuite) TestRollbackOnErrorInSideEffect() {
	var count int
	var err error
	var sideEffectTestError = errors.New("some error occurred in side-effect")
	var createSideEffect = func() error {
		return sideEffectTestError
	}
	const user = "zazie"

	var iconfile1 = createTestIconfile("metro-icon", "french", "great")
	err = repositories.CreateIcon(getPool(), iconfile1, user, nil)
	s.NoError(err)

	var iconfile2 = createTestIconfile("animal-icon", "french", "huge")
	err = repositories.CreateIcon(getPool(), iconfile2, user, createSideEffect)
	s.True(errors.Is(err, sideEffectTestError))

	count, err = getIconCount()
	s.NoError(err)
	s.Equal(1, count)
	s.getIconfileChecked(iconfile1)
	_, err = s.getIconfile(iconfile2)
	s.True(errors.Is(err, domain.ErrIconNotFound))
}
