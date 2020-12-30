package itests

import (
	"github.com/pdkovacs/igo-repo/backend/pkg/icon"
)

type testIconDescriptor struct {
	name       string
	modifiedBy string
	files      []icon.IconfileDescriptor
}

var testIconInputDataDescriptor = []testIconDescriptor{
	{
		name:       "attach_money",
		modifiedBy: "ux",
		files: []icon.IconfileDescriptor{
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
		files: []icon.IconfileDescriptor{
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
