package db

import (
	"crypto/rand"
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

func (s *addIconToDBTestSuite) getCheckIconfile(iconfile domain.Iconfile) {
	content, err := repositories.GetIconFile(getPool(), iconfile.Name, iconfile.Format, iconfile.Size)
	s.NoError(err)
	s.Equal(iconfile.Content, content)
}

func (s *addIconToDBTestSuite) TestAddFirstIcon() {
	const user = "zazie"
	var iconfile = createTestIconfile("metro-icon", "french", "great")
	fmt.Printf("Hello, First Icon %v\n", iconfile.Name)
	repositories.CreateIcon(getPool(), iconfile, user)
	s.getCheckIconfile(iconfile)
}

func (s *addIconToDBTestSuite) TestAddASecondIcon() {
	const user = "zazie"
	var iconfile1 = createTestIconfile("metro-icon", "french", "great")
	var iconfile2 = createTestIconfile("animal-icon", "french", "huge")
	repositories.CreateIcon(getPool(), iconfile1, user)
	repositories.CreateIcon(getPool(), iconfile2, user)
	count, err := getIconCount()
	s.NoError(err)
	s.Equal(2, count)
	s.getCheckIconfile(iconfile1)
	s.getCheckIconfile(iconfile2)
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func createTestIconfile(name, format, size string) domain.Iconfile {
	return domain.Iconfile{
		IconAttributes: domain.IconAttributes{
			Name: name,
		},
		IconfileData: domain.IconfileData{
			IconfileDescriptor: domain.IconfileDescriptor{
				Format: format,
				Size:   size,
			},
			Content: randomBytes(4096),
		},
	}
}
