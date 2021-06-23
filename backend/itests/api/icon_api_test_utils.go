package api

import (
	"fmt"
	"os"
	"path"

	"github.com/pdkovacs/igo-repo/backend/itests/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

var backendSourceHome = os.Getenv("BACKEND_SOURCE_HOME")

func init() {
	if backendSourceHome == "" {
		backendSourceHome = fmt.Sprintf("%s/github/pdkovacs/igo-repo/backend", os.Getenv("HOME"))
	}
}

func getDemoIconfileContent(iconName string, iconfile domain.IconfileDescriptor) []byte {
	pathToContent := path.Join(backendSourceHome, "demo-data", iconfile.Format, iconfile.Size, fmt.Sprintf("%s.%s", iconName, iconfile.Format))
	content, err := os.ReadFile(pathToContent)
	if err != nil {
		panic(err)
	}
	return content
}

type testIconDescriptor struct {
	name       string
	modifiedBy string
	files      []domain.IconfileDescriptor
}

var testIconInputDataDescriptor = []testIconDescriptor{
	{
		name:       "attach_money",
		modifiedBy: "ux",
		files: []domain.IconfileDescriptor{
			{
				Format: "svg",
				Size:   "18px",
			},
			{
				Format: "svg",
				Size:   "24px",
			},
			{
				Format: "png",
				Size:   "24dp",
			},
		},
	},
	{
		name:       "cast_connected",
		modifiedBy: defaultCredentials.Username,
		files: []domain.IconfileDescriptor{
			{
				Format: "svg",
				Size:   "24px",
			},
			{
				Format: "svg",
				Size:   "48px",
			},
			{
				Format: "png",
				Size:   "24dp",
			},
		},
	},
}

var moreTestIconInputDataDescriptor = []testIconDescriptor{
	{
		name:       "format_clear",
		modifiedBy: defaultCredentials.Username,
		files: []domain.IconfileDescriptor{
			{
				Format: "png",
				Size:   "24dp",
			},
			{
				Format: "svg",
				Size:   "48px",
			},
		},
	},
	{
		name:       "insert_photo",
		modifiedBy: defaultCredentials.Username,
		files: []domain.IconfileDescriptor{
			{
				Format: "png",
				Size:   "24dp",
			},
			{
				Format: "svg",
				Size:   "48px",
			},
		},
	},
}

var dp2px = map[string]string{
	"24dp": "36px",
	"36dp": "54px",
}

type ingestedIconfileData struct {
	format string
	size   string
	path   string
}

type ingestedIconDataDescription struct {
	name       string
	modifiedBy string
	paths      []ingestedIconfileData
	tags       []string
}

var ingestedTestIconDataDescription = []ingestedIconDataDescription{
	{
		name:       "attach_money",
		modifiedBy: "ux",
		paths: []ingestedIconfileData{
			{format: "png", size: "36px", path: "/icon/attach_money/format/png/size/36px"},
			{format: "svg", size: "18px", path: "/icon/attach_money/format/svg/size/18px"},
			{format: "svg", size: "24px", path: "/icon/attach_money/format/svg/size/24px"},
		},
		tags: []string{},
	},
	{
		name:       "cast_connected",
		modifiedBy: "ux",
		paths: []ingestedIconfileData{
			{format: "png", size: "36px", path: "/icon/cast_connected/format/png/size/36px"},
			{format: "svg", size: "24px", path: "/icon/cast_connected/format/svg/size/24px"},
			{format: "svg", size: "48px", path: "/icon/cast_connected/format/svg/size/48px"},
		},
		tags: []string{},
	},
}

func createTestIconInputData(descriptors []testIconDescriptor) []domain.Icon {
	var icons = []domain.Icon{}

	for _, descriptor := range descriptors {

		var iconfiles = []domain.Iconfile{}

		for _, file := range descriptor.files {
			iconfile := createIconfile(file, getDemoIconfileContent(descriptor.name, file))
			iconfiles = append(iconfiles, iconfile)
		}

		icon := domain.Icon{
			IconAttributes: domain.IconAttributes{
				Name:       descriptor.name,
				ModifiedBy: descriptor.modifiedBy,
			},
			Iconfiles: iconfiles,
		}
		icons = append(icons, icon)
	}

	return icons
}

var testIconInputData []domain.Icon

func init() {
	testIconInputData = createTestIconInputData(testIconInputDataDescriptor)
}

func mustAddTestData(session *apiTestSession, testData *[]domain.Icon) {
	var err error
	for _, testIcon := range *testData {
		_, _, err = session.createIcon(testIcon.Name, testIcon.Iconfiles[0].Content)
		if err != nil {
			panic(err)
		}
		for i := 1; i < len(testIcon.Iconfiles); i++ {
			_, _, err = session.addIconfile(testIcon.Name, testIcon.Iconfiles[i])
			if err != nil {
				panic(err)
			}
		}
	}
}

func createIconfile(desc domain.IconfileDescriptor, content []byte) domain.Iconfile {
	return domain.Iconfile{
		IconfileDescriptor: desc,
		Content:            content,
	}
}

type iconTestSuite struct {
	apiTestSuite
}

func (s *iconTestSuite) getCheckIconfile(session apiTestSession, iconName string, iconfile domain.Iconfile) {
	iconfileContent, err := session.GetIconfile(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal(iconfile.Content, iconfileContent)
}

func (s *iconTestSuite) assertGitCleanStatus() {
	repositories.AssertGitCleanStatus(&s.Suite, s.server.Repositories.Git)
}
