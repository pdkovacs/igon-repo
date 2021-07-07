package common

import (
	"crypto/rand"
	"encoding/json"

	"github.com/pdkovacs/igo-repo/internal/auxiliaries"
	"github.com/pdkovacs/igo-repo/internal/domain"
)

func createTestIconfile(format, size string) domain.Iconfile {
	return domain.Iconfile{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: format,
			Size:   size,
		},
		Content: randomBytes(4096),
	}
}

func cloneIconfile(iconfile domain.Iconfile) domain.Iconfile {
	var contentClone = make([]byte, len(iconfile.Content))
	copy(contentClone, iconfile.Content)
	return domain.Iconfile{
		IconfileDescriptor: domain.IconfileDescriptor{
			Format: iconfile.Format,
			Size:   iconfile.Size,
		},
		Content: contentClone,
	}
}

func CloneIcon(icon domain.Icon) domain.Icon {
	var iconfilesClone = make([]domain.Iconfile, len(icon.Iconfiles))
	for _, iconfile := range icon.Iconfiles {
		iconfilesClone = append(iconfilesClone, cloneIconfile(iconfile))
	}
	return domain.Icon{
		IconAttributes: domain.IconAttributes{
			Name:       icon.Name,
			ModifiedBy: icon.ModifiedBy,
			Tags:       icon.Tags,
		},
		Iconfiles: iconfilesClone,
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

func CloneConfig(config auxiliaries.Options) auxiliaries.Options {
	var clone auxiliaries.Options
	var err error

	var configAsJSON []byte
	configAsJSON, err = json.Marshal(config)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(configAsJSON), &clone)
	if err != nil {
		panic(err)
	}

	return clone
}

var TestData = []domain.Icon{
	{
		IconAttributes: domain.IconAttributes{
			Name:       "metro-zazie",
			ModifiedBy: "ux",
			Tags: []string{
				"used-in-marvinjs",
				"some other tag",
			},
		},
		Iconfiles: []domain.Iconfile{
			createTestIconfile("french", "great"),
			createTestIconfile("french", "huge"),
		},
	},
	{
		IconAttributes: domain.IconAttributes{
			Name:       "zazie-icon",
			ModifiedBy: "ux",
			Tags: []string{
				"used-in-marvinjs",
				"yet another tag",
			},
		},
		Iconfiles: []domain.Iconfile{
			createTestIconfile("french", "great"),
			createTestIconfile("dutch", "cute"),
		},
	},
}
