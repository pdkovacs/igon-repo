package domain

import "fmt"

type IconfileDescriptor struct {
	Format string `json:"format"`
	Size   string `json:"size"`
}

// Iconfile the file representation of an icon
type Iconfile struct {
	IconfileDescriptor
	Content []byte
}

func (i Iconfile) String() string {
	return fmt.Sprintf("Format: %s, Size: %s, Content: [%d bytes long]", i.Format, i.Size, len(i.Content))
}

type IconAttributes struct {
	Name       string
	ModifiedBy string
	Tags       []string
}

type IconDescriptor struct {
	IconAttributes
	Iconfiles []IconfileDescriptor
}

// IconDescriptor describes an icon
type Icon struct {
	IconAttributes
	Iconfiles []Iconfile
}
