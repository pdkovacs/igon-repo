package repositories

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

// func describeIconInTx(tx *sql.Tx, iconName string, forUpdate bool) (domain.IconDescriptor, error) {
// 	var err error
// 	var rows *sql.Rows

// 	var forUpdateClause = ""
// 	if forUpdate {
// 		forUpdateClause = " FOR UPDATE"
// 	}
// 	var iconSQL = "SELECT id, modified_by FROM icon WHERE name = $1" + forUpdateClause
// 	var iconfilesSQL = "SELECT file_format, icon_size FROM icon_file " +
// 		"WHERE icon_id = $1 " +
// 		"ORDER BY file_format, icon_size" + forUpdateClause
// 	var tagsSQL = "SELECT text FROM tag, icon_to_tags " +
// 		"WHERE icon_to_tags.icon_id = $1 " +
// 		"AND icon_to_tags.tag_id = tag.id" + forUpdateClause

// 	var iconId int
// 	var modifiedBy string
// 	err = tx.QueryRow(iconSQL, iconName).Scan(&iconId, &modifiedBy)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return domain.IconDescriptor{}, domain.ErrIconNotFound
// 		} else {
// 			return domain.IconDescriptor{}, fmt.Errorf("error while retrieving icon \"%s\" from database: %w", iconName, err)
// 		}
// 	}

// 	iconfiles := make([]domain.Iconfile, 0, 10)
// 	tags := make([]string, 0, 50)

// 	rows, err = tx.Query(iconfilesSQL, iconId)
// 	if err != nil {
// 		return domain.IconDescriptor{}, fmt.Errorf("error while retrieving iconfiles for \"%s\" from database: %w", iconName, err)
// 	}
// 	defer rows.Close()
// 	var format string
// 	var size string
// 	for rows.Next() {
// 		err = rows.Scan(&format, &size)
// 		if err != nil {
// 			return domain.IconDescriptor{}, fmt.Errorf("error while retrieving iconfiles for \"%s\" from database: %w", iconName, err)
// 		}
// 		append(iconfiles, domain.Iconfile{

// 		})
// 	}
// 		map(iconfileResult => iconfileResult.rows.reduce(
// 				(icon: IconDescriptor, row: any) => icon.addIconfile({
// 					format: row.file_format,
// 					size: row.icon_size
// 				}),
// 				initialIconInfo
// 			)),

// }

// func DescribeIcon(db *sql.DB, iconName string) (domain.IconDescriptor, error) {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return domain.IconDescriptor{}, err
// 	}
// 	defer tx.Rollback()
// 	return describeIconInTx(tx, iconName, false)
// }

type CreateSideEffect func() error

func CreateIcon(db *sql.DB, iconName string, iconfile domain.Iconfile, modifiedBy string, createSideEffect CreateSideEffect) error {
	var tx *sql.Tx
	var err error

	tx, err = db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertIconSQL string = "INSERT INTO icon(name, modified_by) " +
		"VALUES($1, $2) RETURNING id"
	_, err = tx.Exec(insertIconSQL, iconName, modifiedBy)
	if err != nil {
		return err
	}

	err = insertIconfile(tx, iconName, iconfile, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to create icon-file %v: %w", iconName, err)
	}

	if createSideEffect != nil {
		err = createSideEffect()
		if err != nil {
			return fmt.Errorf("failed to create icon file %s, %w", iconName, err)
		}
	}

	tx.Commit()
	return nil
}

func AddIconfileToIcon(db *sql.DB, iconName string, iconfile domain.Iconfile, modifiedBy string, createSideEffect CreateSideEffect) error {
	var tx *sql.Tx
	var err error

	tx, err = db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = insertIconfile(tx, iconName, iconfile, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to create icon-file %v: %w", iconName, err)
	}

	if createSideEffect != nil {
		err = createSideEffect()
		if err != nil {
			return fmt.Errorf("failed to create icon file %s, %w", iconName, err)
		}
	}

	tx.Commit()
	return nil
}

func insertIconfile(tx *sql.Tx, iconName string, iconfile domain.Iconfile, modifiedBy string) error {
	const insertIconfileSQL = "INSERT INTO icon_file(icon_id, file_format, icon_size, content) " +
		"SELECT id, $2, $3, $4 FROM icon WHERE name = $1 RETURNING id"
	_, err := tx.Exec(insertIconfileSQL, iconName, iconfile.Format, iconfile.Size, iconfile.Content)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				return domain.ErrIconfileAlreadyExists
			}
		}
		return fmt.Errorf("failed to insert icon-file %v: %w", iconName, err)
	}
	return nil
}

func GetIconFile(db *sql.DB, iconName, format, iconSize string) ([]byte, error) {
	const getIconfileSQL = "SELECT content FROM icon, icon_file " +
		"WHERE icon_id = icon.id AND " +
		"file_format = $2 AND " +
		"icon_size = $3 AND " +
		"icon.name = $1"

	var rows *sql.Rows
	var err error
	rows, err = db.Query(getIconfileSQL, iconName, format, iconSize)
	if err != nil {
		return []byte{}, err
	}
	defer rows.Close()

	var content = []byte{}
	if !rows.Next() {
		return content, domain.ErrIconNotFound
	}
	err = rows.Scan(&content)
	if err != nil {
		return []byte{}, err
	}
	if rows.Next() {
		return []byte{}, domain.ErrTooManyIconsFound
	}
	err = rows.Err()
	if err != nil {
		return []byte{}, err
	}

	return content, nil
}

func GetExistingTags(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT text FROM tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]string, 0, 50)
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tags, nil
}
