package indexing

import (
	"fmt"
	"os"

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

type IndexRepoTestExtension interface {
	Close() error
	GetIconCount() (int, error)
	GetIconFileCount() (int, error)
	GetTagRelationCount() (int, error)
	ResetData() error
}

type TestIndexRepository interface {
	repositories.IndexRepository
	IndexRepoTestExtension
}

type TestIndexRepositoryFactory func(conf *config.Options) (TestIndexRepository, error)

type IndexTestRepoController struct {
	repo        TestIndexRepository
	repoFactory TestIndexRepositoryFactory
}

func (ctl *IndexTestRepoController) ResetRepo(conf *config.Options) error {
	if ctl.repo == nil {
		var factoryErr error
		ctl.repo, factoryErr = ctl.repoFactory(conf)
		if factoryErr != nil {
			return factoryErr
		}
	}
	return ctl.repo.ResetData()
}

func (ctl *IndexTestRepoController) Close() error {
	return ctl.repo.Close()
}

func (ctl *IndexTestRepoController) GetIconCount() (int, error) {
	return ctl.repo.GetIconCount()
}

func (ctl *IndexTestRepoController) DescribeIcon(iconName string) (domain.IconDescriptor, error) {
	return ctl.repo.DescribeIcon(iconName)
}

func (ctl *IndexTestRepoController) DescribeAllIcons() ([]domain.IconDescriptor, error) {
	return ctl.repo.DescribeAllIcons()
}

func (ctl *IndexTestRepoController) CreateIcon(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.CreateIcon(iconName, iconfile, modifiedBy, createSideEffect)
}

func (ctl *IndexTestRepoController) AddIconfileToIcon(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.AddIconfileToIcon(iconName, iconfile, modifiedBy, createSideEffect)
}

func (ctl *IndexTestRepoController) AddTag(iconName string, tag string, modifiedBy string) error {
	return ctl.repo.AddTag(iconName, tag, modifiedBy)
}

func (ctl *IndexTestRepoController) GetExistingTags() ([]string, error) {
	return ctl.repo.GetExistingTags()
}

func (ctl *IndexTestRepoController) DeleteIcon(iconName string, modifiedBy string) error {
	return ctl.repo.DeleteIcon(iconName, modifiedBy, nil)
}

func (ctl *IndexTestRepoController) GetIconFileCount() (int, error) {
	return ctl.repo.GetIconFileCount()
}

func (ctl *IndexTestRepoController) GetTagRelationCount() (int, error) {
	return ctl.repo.GetTagRelationCount()
}

func (ctl *IndexTestRepoController) DeleteIconfile(iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.DeleteIconfile(iconName, iconfile, modifiedBy, nil)
}

func NewTestPgRepo(conf *config.Options) (TestIndexRepository, error) {
	var err error
	connection, err := pgdb.NewDBConnection(*conf)
	if err != nil {
		return nil, err
	}
	_, schemaErr := pgdb.OpenSchema(*conf, connection)
	if schemaErr != nil {
		return nil, schemaErr
	}

	sqlDb := pgdb.NewDBRepository(connection)
	return &PgTestRepository{
		Repository: &sqlDb,
	}, nil
}

type IndexingTestSuite struct {
	*suite.Suite
	config             config.Options
	testRepoController IndexTestRepoController
	logger             zerolog.Logger
}

func (s *IndexingTestSuite) SetupSuite() {
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	s.config = conf
	s.config.DBSchemaName = "itest_repositories"
	s.logger = logging.Get()
	NewTestPgRepo(&conf)
}

func (s *IndexingTestSuite) TearDownSuite() {
	s.testRepoController.Close()
}

func (s *IndexingTestSuite) BeforeTest(suiteName, testName string) {
	err := s.testRepoController.ResetRepo(&s.config)
	if err != nil {
		panic(fmt.Sprintf("failed to delete test data: %v", err))
	}
}

func (s *IndexingTestSuite) equalIconAttributes(icon1 domain.Icon, icon2 domain.IconDescriptor, expectedTags []string) {
	s.Equal(icon1.Name, icon2.Name)
	s.Equal(icon1.ModifiedBy, icon2.ModifiedBy)
	if expectedTags != nil {
		s.Equal(expectedTags, icon2.Tags)
	}
}

func (s *IndexingTestSuite) getIconCount() (int, error) {
	return s.testRepoController.GetIconCount()
}

var DefaultIndexTestRepoController IndexTestRepoController = IndexTestRepoController{
	repoFactory: func(conf *config.Options) (TestIndexRepository, error) {
		return NewTestPgRepo(conf)
	},
}

func IndexProvidersToTest() []IndexTestRepoController {
	if len(os.Getenv("PG_ONLY")) > 0 {
		return []IndexTestRepoController{DefaultIndexTestRepoController}
	}
	return []IndexTestRepoController{
		DefaultIndexTestRepoController,
	}
}

func indexingTestSuites() []IndexingTestSuite {
	all := []IndexingTestSuite{}

	for _, provider := range IndexProvidersToTest() {
		suiteToEmbed := new(suite.Suite)
		conf := test_commons.CloneConfig(test_commons.GetTestConfig())
		all = append(all, IndexingTestSuite{
			suiteToEmbed,
			conf,
			provider,
			logging.Get().With().Str("test_sequence_name", "indexing tests").Logger(),
		})
	}

	return all
}
