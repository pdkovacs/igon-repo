package domain

// IconfileDescriptor describes an icon-file
type IconfileDescriptor struct {
	Format string
	Size   string
}

// IconfileData holds data of an icon-file
type IconfileData struct {
	IconfileDescriptor
	Content []byte
}

// IconAttributes holds attributes of an icon
type IconAttributes struct {
	Name string
}

// Iconfile describes the file representation of an icon
type Iconfile struct {
	IconAttributes
	IconfileData
}

// IconDescriptor describes an icon
type IconDescriptor struct {
	IconAttributes
	Name       string
	ModifiedBy string
	Iconfiles  []IconfileDescriptor
	Tags       []string
}
