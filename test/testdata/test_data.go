package testdata

import (
	"fmt"
	"os"
	"path"

	"igo-repo/internal/app/domain"
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/config"
	httpadapter "igo-repo/internal/http"
)

var backendSourceHome = os.Getenv("BACKEND_SOURCE_HOME")

var DefaultCredentials = config.PasswordCredentials{Username: "ux", Password: "ux"}
var defaultUserID = authn.LocalDomain.CreateUserID(DefaultCredentials.Username)

func init() {
	if backendSourceHome == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
		backendSourceHome = fmt.Sprintf("%s/github/pdkovacs/igo-repo", homeDir)
	}
}

func GetDemoIconfileContent(iconName string, iconfile domain.IconfileDescriptor) []byte {
	pathToContent := path.Join(backendSourceHome, "test/demo-data", iconfile.Format, iconfile.Size, fmt.Sprintf("%s.%s", iconName, iconfile.Format))
	content, err := os.ReadFile(pathToContent)
	if err != nil {
		panic(err)
	}
	return content
}

var testIconInputDataDescriptor = []domain.IconDescriptor{
	{
		IconAttributes: domain.IconAttributes{
			Name:       "attach_money",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
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
		IconAttributes: domain.IconAttributes{
			Name:       "cast_connected",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
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

var moreTestIconInputDataDescriptor = []domain.IconDescriptor{
	{
		IconAttributes: domain.IconAttributes{
			Name:       "format_clear",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
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
		IconAttributes: domain.IconAttributes{
			Name:       "insert_photo",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
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

var DP2PX = map[string]string{
	"24dp": "36px",
	"36dp": "54px",
	"18px": "18px",
	"24px": "24px",
	"48px": "48px",
}

var testResponseIconData []httpadapter.IconDTO

var moreTestResponseIconData []httpadapter.IconDTO

func createTestIconInputData(descriptors []domain.IconDescriptor) []domain.Icon {
	var icons = []domain.Icon{}

	for _, descriptor := range descriptors {

		var iconfiles = []domain.Iconfile{}

		for _, file := range descriptor.Iconfiles {
			iconfile := createIconfile(file, GetDemoIconfileContent(descriptor.Name, file))
			iconfiles = append(iconfiles, iconfile)
		}

		icon := domain.Icon{
			IconAttributes: descriptor.IconAttributes,
			Iconfiles:      iconfiles,
		}
		icons = append(icons, icon)
	}

	return icons
}

var testIconInputDataMaster []domain.Icon
var moreTestIconInputDataMaster []domain.Icon

func mapIconfileSize(iconfile domain.IconfileDescriptor) domain.IconfileDescriptor {
	mappedIconfile := iconfile
	mappedSize, has := DP2PX[iconfile.Size]
	if !has {
		panic(fmt.Sprintf("Icon size %s cannot be mapped", iconfile.Size))
	}
	mappedIconfile.Size = mappedSize
	return mappedIconfile
}

func mapIconfileSizes(iconDescriptor domain.IconDescriptor) domain.IconDescriptor {
	newIconfiles := []domain.IconfileDescriptor{}
	for _, iconfile := range iconDescriptor.Iconfiles {
		newIconfiles = append(newIconfiles, mapIconfileSize(iconfile))
	}
	mappedIcon := iconDescriptor
	mappedIcon.Iconfiles = newIconfiles

	return mappedIcon
}

func init() {
	testIconInputDataMaster = createTestIconInputData(testIconInputDataDescriptor)
	moreTestIconInputDataMaster = createTestIconInputData(moreTestIconInputDataDescriptor)

	testResponseIconData = []httpadapter.IconDTO{}
	for _, testIconDescriptor := range testIconInputDataDescriptor {
		testResponseIconData = append(testResponseIconData, httpadapter.CreateResponseIcon("/icon", mapIconfileSizes(testIconDescriptor)))
	}

	moreTestResponseIconData = []httpadapter.IconDTO{}
	for _, testIconDescriptor := range moreTestIconInputDataDescriptor {
		moreTestResponseIconData = append(moreTestResponseIconData, httpadapter.CreateResponseIcon("/icon", mapIconfileSizes(testIconDescriptor)))
	}
}

func createIconfile(desc domain.IconfileDescriptor, content []byte) domain.Iconfile {
	return domain.Iconfile{
		IconfileDescriptor: desc,
		Content:            content,
	}
}

func getTestData(icons []domain.Icon, responseIcons []httpadapter.IconDTO) ([]domain.Icon, []httpadapter.IconDTO) {
	iconListClone := []domain.Icon{}
	for _, icon := range icons {
		iconfilesClone := make([]domain.Iconfile, len(icon.Iconfiles))
		tagsClone := make([]string, len(icon.Tags))
		copy(iconfilesClone, icon.Iconfiles)
		copy(tagsClone, icon.Tags)
		iconClone := domain.Icon{
			IconAttributes: domain.IconAttributes{
				Name:       icon.Name,
				ModifiedBy: icon.ModifiedBy,
				Tags:       tagsClone,
			},
			Iconfiles: iconfilesClone,
		}
		iconListClone = append(iconListClone, iconClone)
	}

	responseIconListClone := []httpadapter.IconDTO{}
	for _, resp := range responseIcons {
		paths := make([]httpadapter.IconPath, len(resp.Paths))
		tags := make([]string, len(resp.Tags))
		copy(paths, resp.Paths)
		copy(tags, resp.Tags)
		respClone := httpadapter.IconDTO{
			Name:       resp.Name,
			Paths:      paths,
			Tags:       tags,
			ModifiedBy: resp.ModifiedBy,
		}
		responseIconListClone = append(responseIconListClone, respClone)
	}

	return iconListClone, responseIconListClone
}

func Get() ([]domain.Icon, []httpadapter.IconDTO) {
	return getTestData(testIconInputDataMaster, testResponseIconData)
}

func GetMore() ([]domain.Icon, []httpadapter.IconDTO) {
	return getTestData(moreTestIconInputDataMaster, moreTestResponseIconData)
}
