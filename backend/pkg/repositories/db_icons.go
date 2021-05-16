package repositories

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

func describeIconInTx(tx *sql.Tx, iconName string, forUpdate bool) (domain.Icon, error) {
	var err error
	var rows *sql.Rows

	var forUpdateClause = ""
	if forUpdate {
		forUpdateClause = " FOR UPDATE"
	}
	var iconSQL = "SELECT id, modified_by FROM icon WHERE name = $1" + forUpdateClause
	var iconfilesSQL = "SELECT file_format, icon_size FROM icon_file " +
		"WHERE icon_id = $1 " +
		"ORDER BY file_format, icon_size" + forUpdateClause
	var tagsSQL = "SELECT text FROM tag, icon_to_tags " +
		"WHERE icon_to_tags.icon_id = $1 " +
		"AND icon_to_tags.tag_id = tag.id" + forUpdateClause

	var iconId int
	var modifiedBy string
	err = tx.QueryRow(iconSQL, iconName).Scan(&iconId, &modifiedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Icon{}, domain.ErrIconNotFound
		} else {
			return domain.Icon{}, fmt.Errorf("error while retrieving icon '%s' from database: %w", iconName, err)
		}
	}

	iconfiles := make([]domain.Iconfile, 0, 10)
	rows, err = tx.Query(iconfilesSQL, iconId)
	if err != nil {
		return domain.Icon{}, fmt.Errorf("error while retrieving iconfiles for '%s' from database: %w", iconName, err)
	}
	defer rows.Close()
	var format string
	var size string
	for rows.Next() {
		err = rows.Scan(&format, &size)
		if err != nil {
			return domain.Icon{}, fmt.Errorf("error while retrieving iconfiles for '%s' from database: %w", iconName, err)
		}
		iconfiles = append(iconfiles, domain.Iconfile{
			Format: format,
			Size:   size,
		})
	}

	tags := make([]string, 0, 50)
	rows, err = tx.Query(tagsSQL, iconId)
	if err != nil {
		return domain.Icon{}, fmt.Errorf("error while retrieving tags for '%s' from database: %w", iconName, err)
	}
	defer rows.Close()
	var tag string
	for rows.Next() {
		err = rows.Scan(&tag)
		if err != nil {
			return domain.Icon{}, fmt.Errorf("error while retrieving tags for '%s' from database: %w", iconName, err)
		}
		tags = append(tags, tag)
	}

	return domain.Icon{
		Name:      iconName,
		Iconfiles: iconfiles,
		Tags:      tags,
	}, nil
}

func DescribeIcon(db *sql.DB, iconName string) (domain.Icon, error) {
	tx, err := db.Begin()
	if err != nil {
		return domain.Icon{}, err
	}
	defer tx.Rollback()
	return describeIconInTx(tx, iconName, false)
}

type CreateSideEffect func() error

func CreateIcon(db *sql.DB, iconName string, iconfile domain.Iconfile, modifiedBy string, createSideEffect CreateSideEffect) error {
	var tx *sql.Tx
	var err error

	tx, err = db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction when creating icon %v: %w", iconName, err)
	}
	defer tx.Rollback()

	const insertIconSQL string = "INSERT INTO icon(name, modified_by) " +
		"VALUES($1, $2) RETURNING id"
	_, err = tx.Exec(insertIconSQL, iconName, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to create icon %v: %w", iconName, err)
	}

	err = insertIconfile(tx, iconName, iconfile, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to create icon-file for %v: %w", iconName, err)
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
		return fmt.Errorf("failed to start transaction when creating icon-file for %v: %w", iconName, err)
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

	var err error
	var content = []byte{}
	err = db.QueryRow(getIconfileSQL, iconName, format, iconSize).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return content, domain.ErrIconNotFound
		}
		return []byte{}, fmt.Errorf("failed to get iconfile %v: %w", iconName, err)
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
			return nil, fmt.Errorf("failed to retrieve all tags: %w", err)
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all tags: %w", err)
	}

	return tags, nil
}

func createTag(tx *sql.Tx, tag string) (int64, error) {
	var id int64
	// The lib/pq people messed the API up :-( : https://github.com/lib/pq/issues/24#issuecomment-841794798
	err := tx.QueryRow("INSERT INTO tag(text) VALUES($1) RETURNING id", tag).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last inserted id: %w", err)
	}
	return id, nil
}

func addTagReferenceToIcon(tx *sql.Tx, tagId int64, iconName string) error {
	_, err := tx.Exec("INSERT INTO icon_to_tags(icon_id, tag_id) SELECT id, $1 FROM icon WHERE name = $2", tagId, iconName)
	if err != nil {
		return fmt.Errorf("failed to add tag reference %d to icon '%s': %w", tagId, iconName, err)
	}
	return nil
}

func GetTagId(tx *sql.Tx, tag string) (int64, error) {
	var tagId int64
	err := tx.QueryRow("SELECT id FROM tag WHERE text = $1", tag).Scan(&tagId)
	if err != nil {
		if err == sql.ErrNoRows {
			return createTag(tx, tag)
		}
		return 0, err
	}
	return tagId, nil
}

func AddTag(db *sql.DB, iconName string, tag string) error {
	tx, trError := db.Begin()
	if trError != nil {
		return fmt.Errorf("failed to obtain transaction for adding tag '%s' to '%s': %w", tag, iconName, trError)
	}
	defer tx.Rollback()

	tagId, insertTagErr := GetTagId(tx, tag)
	if insertTagErr != nil {
		return fmt.Errorf("failed to insert tag '%s' for '%s': %w", tag, iconName, insertTagErr)
	}
	addRefErr := addTagReferenceToIcon(tx, tagId, iconName)
	if addRefErr != nil {
		return fmt.Errorf("failed to connect tag '%s' to icon '%s': %w", tag, iconName, addRefErr)
	}

	tx.Commit()
	return nil
}
