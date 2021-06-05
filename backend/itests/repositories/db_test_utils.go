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

type DBTestSuite struct {
	suite.Suite
	config auxiliaries.Options
	dbRepo *repositories.DatabaseRepository
}

func DeleteDBData(db *sql.DB) error {
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

	err = tx.Commit()
	return nil
}

func (s *DBTestSuite) NewTestDBRepo() {
	var logger = log.WithField("prefix", "make-sure-has-uptodate-db-schema-with-no-data")
	var err error
	config := auxiliaries.GetDefaultConfiguration()
	s.dbRepo, err = repositories.InitDBRepo(config)
	if err != nil {
		panic(err)
	}
	err = DeleteDBData(s.dbRepo.ConnectionPool)
	if err != nil {
		logger.Errorf("failed to delete test data: %v", err)
		panic(err)
	}
}

func (s DBTestSuite) getIconCount() (int, error) {
	var getIconCountSQL = "SELECT count(*) from icon"
	var count int
	err := s.dbRepo.ConnectionPool.QueryRow(getIconCountSQL).Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func manageTestResourcesAfterEach() {
}

func (s *DBTestSuite) SetupSuite() {
	s.config.DBSchemaName = "itest_repositories"
}

func (s *DBTestSuite) TearDownSuite() {
	s.dbRepo.Close()
}

func (s *DBTestSuite) BeforeTest(suiteName, testName string) {
	s.NewTestDBRepo()
}

func (s *DBTestSuite) AfterTest(suiteName, testName string) {
	manageTestResourcesAfterEach()
}

func (s *DBTestSuite) equalIconAttributes(icon1 domain.Icon, icon2 domain.Icon, expectedTags []string) {
	s.Equal(icon1.Name, icon2.Name)
	s.Equal(icon1.ModifiedBy, icon2.ModifiedBy)
	if expectedTags != nil {
		s.Equal(expectedTags, icon2.Tags)
	}
}

func (s *DBTestSuite) getIconfile(iconName string, iconfile domain.Iconfile) ([]byte, error) {
	return s.dbRepo.GetIconFile(iconName, iconfile.Format, iconfile.Size)
}

func (s *DBTestSuite) getIconfileChecked(iconName string, iconfile domain.Iconfile) {
	content, err := s.getIconfile(iconName, iconfile)
	s.NoError(err)
	s.Equal(iconfile.Content, content)
}
