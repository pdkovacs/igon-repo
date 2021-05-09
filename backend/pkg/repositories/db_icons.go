package repositories

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
)

type CreateSideEffect func() error

func CreateIcon(db *sql.DB, iconfile domain.Iconfile, modifiedBy string, createSideEffect CreateSideEffect) error {
	var tx *sql.Tx
	var err error

	tx, err = db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertIconSQL string = "INSERT INTO icon(name, modified_by) " +
		"VALUES($1, $2) RETURNING id"
	_, err = tx.Exec(insertIconSQL, iconfile.Name, modifiedBy)
	if err != nil {
		return err
	}

	err = insertIconfile(tx, iconfile, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to create icon-file %v: %w", iconfile.Name, err)
	}

	if createSideEffect != nil {
		err = createSideEffect()
		if err != nil {
			return fmt.Errorf("failed to create icon file %s, %w", iconfile.Name, err)
		}
	}

	tx.Commit()
	return nil
}

func AddIconfileToIcon(db *sql.DB, iconfile domain.Iconfile, modifiedBy string, createSideEffect CreateSideEffect) error {
	var tx *sql.Tx
	var err error

	tx, err = db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = insertIconfile(tx, iconfile, modifiedBy)
	if err != nil {
		return fmt.Errorf("failed to create icon-file %v: %w", iconfile.Name, err)
	}

	if createSideEffect != nil {
		err = createSideEffect()
		if err != nil {
			return fmt.Errorf("failed to create icon file %s, %w", iconfile.Name, err)
		}
	}

	tx.Commit()
	return nil
}

func insertIconfile(tx *sql.Tx, iconfile domain.Iconfile, modifiedBy string) error {
	const insertIconfileSQL = "INSERT INTO icon_file(icon_id, file_format, icon_size, content) " +
		"SELECT id, $2, $3, $4 FROM icon WHERE name = $1 RETURNING id"
	_, err := tx.Exec(insertIconfileSQL, iconfile.Name, iconfile.Format, iconfile.Size, iconfile.Content)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "unique_violation" {
				return domain.ErrIconfileAlreadyExists
			}
		}
		return fmt.Errorf("failed to insert icon-file %v: %w", iconfile.Name, err)
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
