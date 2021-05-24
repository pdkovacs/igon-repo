package repositories

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

var db *sql.DB

var testIconRepoSchema = "test_iconrepo"

func createTestDBPool() {
	connProps := repositories.CreateConnectionProperties(auxiliaries.GetDefaultConfiguration())
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&options=-csearch_path=%s",
		connProps.User,
		connProps.Password,
		connProps.Host,
		connProps.Database,
		testIconRepoSchema,
	)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	db.Ping()
}

func terminatePool() {
	db.Close()
}

func deleteData() error {
	var tx *sql.Tx
	var err error

	tx, err = db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start Tx for deleting test data: %w", err)
	}
	defer tx.Rollback()

	tables := []string{"icon", "icon_file", "tag", "icon_to_tags"}
	for _, table := range tables {
		_, err = tx.Exec("DELETE FROM " + table)
		if err != nil {
			return fmt.Errorf("failed to delete test data from table %s: %w", table, err)
		}
	}

	tx.Commit()
	return nil
}

func makeSureHasUptodateDBSchemaWithNoData() {
	var logger = log.WithField("prefix", "make-sure-has-uptodate-db-schema-with-no-data")
	var err error
	err = repositories.CreateSchemaRetry(db, testIconRepoSchema)
	if err != nil {
		logger.Errorf("Failed to create schema %v", err)
		panic(err)
	}
	err = repositories.ExecuteSchemaUpgrade(db)
	if err != nil {
		panic(err)
	}
	err = deleteData()
	if err != nil {
		logger.Errorf("failed to delete test data: %v", err)
		panic(err)
	}
}

func getPool() *sql.DB {
	return db
}

func getIconCount() (int, error) {
	var getIconCountSQL = "SELECT count(*) from icon"
	var count int
	err := db.QueryRow(getIconCountSQL).Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func manageTestResourcesBeforeAll() {
	createTestDBPool()
}

func manageTestResourcesAfterAll() {
	terminatePool()
}

func manageTestResourcesBeforeEach() {
	makeSureHasUptodateDBSchemaWithNoData()
}

func manageTestResourcesAfterEach() {
}

type dbTestSuite struct {
	suite.Suite
}

func (s *dbTestSuite) SetupSuite() {
	manageTestResourcesBeforeAll()
}

func (s *dbTestSuite) TearDownSuite() {
	manageTestResourcesAfterAll()
}

func (s *dbTestSuite) BeforeTest(suiteName, testName string) {
	manageTestResourcesBeforeEach()
}

func (s *dbTestSuite) AfterTest(suiteName, testName string) {
	manageTestResourcesAfterEach()
}

func (s *dbTestSuite) equalIconAttributes(icon1 domain.Icon, icon2 domain.Icon, expectedTags []string) {
	s.Equal(icon1.Name, icon2.Name)
	s.Equal(icon1.ModifiedBy, icon2.ModifiedBy)
	if expectedTags != nil {
		s.Equal(expectedTags, icon2.Tags)
	}
}

func (s *dbTestSuite) getIconfile(iconName string, iconfile domain.Iconfile) ([]byte, error) {
	return repositories.GetIconFile(getPool(), iconName, iconfile.Format, iconfile.Size)
}

func (s *dbTestSuite) getIconfileChecked(iconName string, iconfile domain.Iconfile) {
	content, err := s.getIconfile(iconName, iconfile)
	s.NoError(err)
	s.Equal(iconfile.Content, content)
}
