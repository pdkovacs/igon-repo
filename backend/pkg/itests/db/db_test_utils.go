package db

import (
	"crypto/rand"
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

var testData = []domain.Icon{
	{
		Name:       "metro-zazie",
		ModifiedBy: "ux",
		Iconfiles: []domain.Iconfile{
			createTestIconfile("french", "great"),
			createTestIconfile("french", "huge"),
		},
		Tags: []string{
			"used-in-marvinjs",
			"some other tag",
		},
	},
	{
		Name:       "zazie-icon",
		ModifiedBy: "ux",
		Iconfiles: []domain.Iconfile{
			createTestIconfile("french", "great"),
			createTestIconfile("dutch", "cute"),
		},
		Tags: []string{
			"used-in-marvinjs",
			"yet another tag",
		},
	},
}

func createTestIconfile(format, size string) domain.Iconfile {
	return domain.Iconfile{
		Format:  format,
		Size:    size,
		Content: randomBytes(4096),
	}
}

func cloneIconfile(iconfile domain.Iconfile) domain.Iconfile {
	var contentClone = make([]byte, len(iconfile.Content))
	copy(contentClone, iconfile.Content)
	return domain.Iconfile{
		Format:  iconfile.Format,
		Size:    iconfile.Size,
		Content: contentClone,
	}
}

func cloneIcon(icon domain.Icon) domain.Icon {
	var iconfilesClone = make([]domain.Iconfile, len(icon.Iconfiles))
	for _, iconfile := range icon.Iconfiles {
		iconfilesClone = append(iconfilesClone, cloneIconfile(iconfile))
	}
	return domain.Icon{
		Name:       icon.Name,
		ModifiedBy: icon.ModifiedBy,
		Tags:       icon.Tags,
		Iconfiles:  iconfilesClone,
	}
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
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
