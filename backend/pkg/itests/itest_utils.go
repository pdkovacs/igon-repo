package itests

import (
	"crypto/rand"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

func createTestIconfile(format, size string) domain.Iconfile {
	return domain.Iconfile{
		Format:  format,
		Size:    size,
		Content: randomBytes(4096),
	}
}

func cloneIconfile(iconfile domain.Iconfile) domain.Iconfile {
	var contentClone = make([]byte, len(iconfile.Content))
	copy(contentClone, iconfile.Content)
	return domain.Iconfile{
		Format:  iconfile.Format,
		Size:    iconfile.Size,
		Content: contentClone,
	}
}

func CloneIcon(icon domain.Icon) domain.Icon {
	var iconfilesClone = make([]domain.Iconfile, len(icon.Iconfiles))
	for _, iconfile := range icon.Iconfiles {
		iconfilesClone = append(iconfilesClone, cloneIconfile(iconfile))
	}
	return domain.Icon{
		Name:       icon.Name,
		ModifiedBy: icon.ModifiedBy,
		Tags:       icon.Tags,
		Iconfiles:  iconfilesClone,
	}
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

var TestData = []domain.Icon{
	{
		Name:       "metro-zazie",
		ModifiedBy: "ux",
		Iconfiles: []domain.Iconfile{
			createTestIconfile("french", "great"),
			createTestIconfile("french", "huge"),
		},
		Tags: []string{
			"used-in-marvinjs",
			"some other tag",
		},
	},
	{
		Name:       "zazie-icon",
		ModifiedBy: "ux",
		Iconfiles: []domain.Iconfile{
			createTestIconfile("french", "great"),
			createTestIconfile("dutch", "cute"),
		},
		Tags: []string{
			"used-in-marvinjs",
			"yet another tag",
		},
	},
}
