package indexing_tests

import (
	"database/sql"
	"fmt"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/indexing/pgdb"
	"iconrepo/test/test_commons"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type TestIndexRepository interface {
	repositories.IndexRepository
	Close() error
	GetIconCount() (int, error)
	GetIconFileCount() (int, error)
	GetTagRelationCount() (int, error)
	ResetDBData() error
}

// TODO:
// func DbProvidersToTest() []TestDBRepository {
// 	if len(os.Getenv("PG_ONLY")) > 0 {
// 		return []TestDBRepository{pgdb.Repository{}}
// 	}
// 	return []GitTestRepo{
// 		blobstore.Local{},
// 		blobstore.Gitlab{},
// 	}
// }

type PostgresTestRepository struct {
	*pgdb.Repository
}

func (sqlDb *PostgresTestRepository) Close() error {
	return sqlDb.Conn.Pool.Close()
}

func (sqlDb *PostgresTestRepository) GetIconCount() (int, error) {
	var rowCount int
	err := sqlDb.Conn.Pool.QueryRow("select count(*) as row_count from icon").Scan(&rowCount)
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}

func (sqlDb *PostgresTestRepository) GetIconFileCount() (int, error) {
	var rowCount int
	err := sqlDb.Conn.Pool.QueryRow("select count(*) as row_count from icon_file").Scan(&rowCount)
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}

func (sqlDb *PostgresTestRepository) GetTagRelationCount() (int, error) {
	var rowCount int
	err := sqlDb.Conn.Pool.QueryRow("select count(*) as row_count from icon_to_tags").Scan(&rowCount)
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}

func (sqlDb *PostgresTestRepository) ResetDBData() error {
	var tx *sql.Tx
	var err error

	tx, err = sqlDb.Conn.Pool.Begin()
	if err != nil {
		return fmt.Errorf("failed to start Tx for deleting test data: %w", err)
	}
	defer tx.Rollback()

	tables := []string{"icon", "icon_file", "tag", "icon_to_tags"}
	for _, table := range tables {
		_, err = tx.Exec("DELETE FROM " + table)
		if err != nil {
			if pgdb.IsDBError(err, pgdb.ErrMissingDBTable) {
				continue
			}
			return fmt.Errorf("failed to delete test data from table %s: %w", table, err)
		}
	}

	_ = tx.Commit()
	return nil
}

func NewTestDbRepositoryFromSQLDB(sqlDb *pgdb.Repository) TestIndexRepository {
	return &PostgresTestRepository{
		Repository: sqlDb,
	}
}

type DBTestSuite struct {
	suite.Suite
	config config.Options
	dbRepo TestIndexRepository
	logger zerolog.Logger
}

func (s *DBTestSuite) NewTestDBRepo() {
	var err error
	config := test_commons.GetTestConfig()
	connection, err := pgdb.NewDBConnection(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create test connection: %v", err))
	}
	_, schemaErr := pgdb.OpenSchema(config, connection)
	if schemaErr != nil {
		panic(schemaErr)
	}

	sqlDb := pgdb.NewDBRepository(connection)
	s.dbRepo = NewTestDbRepositoryFromSQLDB(&sqlDb)
	if err != nil {
		panic(err)
	}
}

func (s *DBTestSuite) SetupSuite() {
	s.config.DBSchemaName = "itest_repositories"
	s.logger = logging.Get()
	s.NewTestDBRepo()
}

func (s *DBTestSuite) TearDownSuite() {
	s.dbRepo.Close()
}

func (s *DBTestSuite) BeforeTest(suiteName, testName string) {
	err := s.dbRepo.ResetDBData()
	if err != nil {
		panic(fmt.Sprintf("failed to delete test data: %v", err))
	}
}

func (s *DBTestSuite) equalIconAttributes(icon1 domain.Icon, icon2 domain.IconDescriptor, expectedTags []string) {
	s.Equal(icon1.Name, icon2.Name)
	s.Equal(icon1.ModifiedBy, icon2.ModifiedBy)
	if expectedTags != nil {
		s.Equal(expectedTags, icon2.Tags)
	}
}

func (s *DBTestSuite) getIconCount() (int, error) {
	return s.dbRepo.GetIconCount()
}
