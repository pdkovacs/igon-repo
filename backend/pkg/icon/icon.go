package icon

// IconfileDescriptor describes an icon-file
type IconfileDescriptor struct {
	Format string
	Size   string
}

// IconfileData holds data of an icon-file
type IconfileData struct {
	IconfileDescriptor
	content []byte
}

// Attributes holds attributes of an icon
type Attributes struct {
	Name string
}

// Iconfile describes the file representation of an icon
type Iconfile struct {
	Attributes
	IconfileData
}

// Descriptor describes an icon
type Descriptor struct {
	Attributes
	Name       string
	ModifiedBy string
	Iconfiles  []IconfileDescriptor
	Tags       []string
}
