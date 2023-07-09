package httpadapter

import (
	"fmt"
	"testing"

	"iconrepo/internal/app/domain"

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

	expectedResponse := IconDTO{
		Name:       name,
		ModifiedBy: modifiedBy,
		Paths: []IconPath{
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "french",
					Size:   "great",
				},
				Path: fmt.Sprintf("%s/cartouche/format/french/size/great", iconPathRoot)},
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "french",
					Size:   "huge",
				},
				Path: fmt.Sprintf("%s/cartouche/format/french/size/huge", iconPathRoot)},
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "english",
					Size:   "OK",
				},
				Path: fmt.Sprintf("%s/cartouche/format/english/size/OK", iconPathRoot)},
			{
				IconfileDescriptor: domain.IconfileDescriptor{
					Format: "english",
					Size:   "nice",
				},
				Path: fmt.Sprintf("%s/cartouche/format/english/size/nice", iconPathRoot)},
		},
		Tags: []string{},
	}

	s.Equal(expectedResponse, CreateResponseIcon(iconPathRoot, iconDescriptor))
}
