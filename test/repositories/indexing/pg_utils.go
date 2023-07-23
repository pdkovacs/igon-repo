package indexing

import (
	"context"
	"database/sql"
	"fmt"
	"iconrepo/internal/repositories/indexing"
	"iconrepo/internal/repositories/indexing/pgdb"
)

type PgTestRepository struct {
	*pgdb.PgRepository
}

func (pgTestRepo *PgTestRepository) Close() error {
	return pgTestRepo.Conn.Pool.Close()
}

func (pgTestRepo *PgTestRepository) GetIconCount(ctx context.Context) (int, error) {
	var rowCount int
	err := pgTestRepo.Conn.Pool.QueryRow("select count(*) as row_count from icon").Scan(&rowCount)
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}

func (pgTestRepo *PgTestRepository) GetIconFileCount(ctx context.Context) (int, error) {
	var rowCount int
	err := pgTestRepo.Conn.Pool.QueryRow("select count(*) as row_count from icon_file").Scan(&rowCount)
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}

func (pgTestRepo *PgTestRepository) GetTagRelationCount(ctx context.Context) (int, error) {
	var rowCount int
	err := pgTestRepo.Conn.Pool.QueryRow("select count(*) as row_count from icon_to_tags").Scan(&rowCount)
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}

func (pgTestRepo *PgTestRepository) ResetData(ctx context.Context) error {
	var tx *sql.Tx
	var err error

	tx, err = pgTestRepo.Conn.Pool.Begin()
	if err != nil {
		return fmt.Errorf("failed to start Tx for deleting test data: %w", err)
	}
	defer tx.Rollback()

	tables := []string{"icon", "icon_file", "tag", "icon_to_tags"}
	for _, table := range tables {
		_, err = tx.Exec("DELETE FROM " + table)
		if err != nil {
			if pgdb.IsDBError(err, indexing.ErrTableNotFound) {
				continue
			}
			return fmt.Errorf("failed to delete test data from table %s: %w", table, err)
		}
	}

	_ = tx.Commit()
	return nil
}
