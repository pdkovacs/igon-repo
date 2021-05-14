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

func createTestDBPool() {
	connProps := repositories.CreateConnectionProperties(auxiliaries.GetDefaultConfiguration())
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", connProps.User, connProps.Password, connProps.Host, connProps.Database)
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

func makeSureHasUptodateDBSchemaWithNoData() {
	var logger = log.WithField("prefix", "make-sure-has-uptodate-db-schema-with-no-data")
	var err error
	err = repositories.CreateSchemaRetry(db)
	if err != nil {
		logger.Errorf("Failed to create schema %v", err)
		panic(err)
	}
	err = repositories.ExecuteSchemaUpgrade(db)
	if err != nil {
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
}
