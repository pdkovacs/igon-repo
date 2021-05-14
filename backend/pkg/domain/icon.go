package domain

// Iconfile describes the file representation of an icon
type Iconfile struct {
	Format  string
	Size    string
	Content []byte
}

// IconDescriptor describes an icon
type Icon struct {
	Name       string
	ModifiedBy string
	Iconfiles  []Iconfile
	Tags       []string
}
