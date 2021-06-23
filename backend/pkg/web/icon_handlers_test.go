package web

import (
	"fmt"
	"testing"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/stretchr/testify/suite"
)

type iconHandlerTestSuite struct {
	suite.Suite
}

func TestIconHandlerTestSuite(t *testing.T) {
	suite.Run(t, &iconHandlerTestSuite{})
}

func (s *iconHandlerTestSuite) TestReturnIconsWithProperPaths() {
	iconPathRoot := "/icon"

	name := "cartouche"
	modifiedBy := "zazie"
	iconfileDescs := []domain.IconfileDescriptor{
		{Format: "french", Size: "great"},
		{Format: "french", Size: "huge"},
		{Format: "english", Size: "OK"},
		{Format: "english", Size: "nice"},
	}
	iconDescriptor := domain.IconDescriptor{
		IconAttributes: domain.IconAttributes{
			Name:       name,
			ModifiedBy: modifiedBy,
			Tags:       []string{},
		},
		Iconfiles: iconfileDescs,
	}

	expectedDTO := IconDTO{
		name:       name,
		modifiedBy: modifiedBy,
		paths: []IconPath{
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "french",
					Size:   "great",
				},
				path: fmt.Sprintf("%s/cartouche/format/french/size/great", iconPathRoot)},
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "french",
					Size:   "huge",
				},
				path: fmt.Sprintf("%s/cartouche/format/french/size/huge", iconPathRoot)},
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "english",
					Size:   "OK",
				},
				path: fmt.Sprintf("%s/cartouche/format/english/size/OK", iconPathRoot)},
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "english",
					Size:   "nice",
				},
				path: fmt.Sprintf("%s/cartouche/format/english/size/nice", iconPathRoot)},
		},
		tags: []string{},
	}

	s.Equal(expectedDTO, createIconDTO(iconPathRoot, iconDescriptor))
}
