package api

import (
	"fmt"
	"os"
	"path"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

var backendSourceHome = os.Getenv("BACKEND_SOURCE_HOME")

func init() {
	if backendSourceHome == "" {
		backendSourceHome = fmt.Sprintf("%s/github/pdkovacs/igo-repo/backend", os.Getenv("HOME"))
	}
}

func getDemoIconfileContent(iconName string, iconfile domain.Iconfile) []byte {
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
	files      []domain.Iconfile
}

var testIconInputDataDescriptor = []testIconDescriptor{
	{
		name:       "attach_money",
		modifiedBy: "ux",
		files: []domain.Iconfile{
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
		files: []domain.Iconfile{
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
