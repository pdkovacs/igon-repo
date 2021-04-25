package repositories

type columnsDefinition map[string]string

type tableSpec struct {
	tableName       string
	columns         columnsDefinition
	col_constraints []string
}

var iconTableColumns = columnsDefinition{
	"id":          "serial primary key",
	"name":        "text",
	"modified_by": "text",
	"modified_at": "timestamp DEFAULT now()",
}

var iconTableSpec = tableSpec{
	tableName: "icon",
	columns:   iconTableColumns,
	col_constraints: []string{
		"UNIQUE (name)",
	},
}

var iconfileTableColumns = columnsDefinition{
	"id":          "serial primary key",
	"icon_id":     "int REFERENCES icon(id) ON DELETE CASCADE",
	"file_format": "text",
	"icon_size":   "text",
	"content":     "bytea",
}

var iconfileTableSpec = tableSpec{
	tableName: "icon_file",
	columns:   iconfileTableColumns,
	col_constraints: []string{
		"UNIQUE (icon_id, file_format, icon_size)",
	},
}

var tagTableSpec = tableSpec{
	tableName: "tag",
	columns: columnsDefinition{
		"id":   "serial primary key",
		"text": "text",
	},
	col_constraints: []string{
		"UNIQUE (text)",
	},
}

var iconToTagsTableSpec = tableSpec{
	tableName: "icon_to_tags",
	columns: columnsDefinition{
		"icon_id": "int REFERENCES icon(id) ON DELETE CASCADE",
		"tag_id":  "int REFERENCES tag(id) ON DELETE CASCADE",
	},
}
