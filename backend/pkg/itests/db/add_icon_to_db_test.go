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

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func (s *addIconToDBTestSuite) getCheckIconfile(iconfile domain.Iconfile) {
	content, err := repositories.GetIconFile(getPool(), iconfile.Name, iconfile.Format, iconfile.Size)
	s.NoError(err)
	s.Equal(iconfile.Content, content)
}

func (s *addIconToDBTestSuite) TestAddFirstIcon() {
	const user = "zazie"
	var iconfile = domain.Iconfile{
		IconAttributes: domain.IconAttributes{
			Name: "metro-icon",
		},
		IconfileData: domain.IconfileData{
			IconfileDescriptor: domain.IconfileDescriptor{
				Format: "french",
				Size:   "great",
			},
			Content: randomBytes(4096),
		},
	}
	fmt.Printf("Hello, First Icon %v\n", iconfile.Name)
	repositories.CreateIcon(getPool(), iconfile, user)
	s.getCheckIconfile(iconfile)
}
