package itests

import (
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

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
		modifiedBy: defaultCredentials.User,
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
