package domain

import "fmt"

// Iconfile describes the file representation of an icon
type Iconfile struct {
	Format  string
	Size    string
	Content []byte
}

func (i Iconfile) String() string {
	return fmt.Sprintf("Format: %s, Size: %s, Content: [%d bytes long]", i.Format, i.Size, len(i.Content))
}

// IconDescriptor describes an icon
type Icon struct {
	Name       string
	ModifiedBy string
	Iconfiles  []Iconfile
	Tags       []string
}
