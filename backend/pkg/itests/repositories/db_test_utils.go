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

type dbTestSuite struct {
	suite.Suite
	*repositories.DatabaseRepository
}

var testIconRepoSchema = "test_iconrepo"

func (s *dbTestSuite) deleteData() error {
	var tx *sql.Tx
	var err error

	tx, err = s.ConnectionPool.Begin()
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

func (s *dbTestSuite) makeSureHasUptodateDBSchemaWithNoData() {
	var logger = log.WithField("prefix", "make-sure-has-uptodate-db-schema-with-no-data")
	var err error
	connProps := repositories.CreateConnectionProperties(auxiliaries.GetDefaultConfiguration())
	repo, errNewDB := repositories.NewDB(connProps, testIconRepoSchema)
	if errNewDB != nil {
		logger.Errorf("Failed to create schema %v", errNewDB)
		panic(errNewDB)
	}
	s.DatabaseRepository = repo
	err = repo.ExecuteSchemaUpgrade()
	if err != nil {
		panic(err)
	}
	err = s.deleteData()
	if err != nil {
		logger.Errorf("failed to delete test data: %v", err)
		panic(err)
	}
}

func (s dbTestSuite) getIconCount() (int, error) {
	var getIconCountSQL = "SELECT count(*) from icon"
	var count int
	err := s.DatabaseRepository.ConnectionPool.QueryRow(getIconCountSQL).Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func manageTestResourcesAfterEach() {
}

func (s *dbTestSuite) SetupSuite() {

}

func (s *dbTestSuite) TearDownSuite() {
	s.DatabaseRepository.Close()
}

func (s *dbTestSuite) BeforeTest(suiteName, testName string) {
	s.makeSureHasUptodateDBSchemaWithNoData()
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
	return s.GetIconFile(iconName, iconfile.Format, iconfile.Size)
}

func (s *dbTestSuite) getIconfileChecked(iconName string, iconfile domain.Iconfile) {
	content, err := s.getIconfile(iconName, iconfile)
	s.NoError(err)
	s.Equal(iconfile.Content, content)
}
