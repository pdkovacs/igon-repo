package indexing

import (
	"context"
	"fmt"
	"os"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/indexing/dynamodb"
	"iconrepo/internal/repositories/indexing/pgdb"
	"iconrepo/test/test_commons"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/suite"
)

type IndexRepoTestExtension interface {
	Close() error
	GetIconCount(ctx context.Context) (int, error)
	GetIconFileCount(ctx context.Context) (int, error)
	GetTagRelationCount(ctx context.Context) (int, error)
	ResetData(ctx context.Context) error
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

func (ctl *IndexTestRepoController) ResetRepo(ctx context.Context, conf *config.Options) error {
	if ctl.repo == nil {
		var factoryErr error
		ctl.repo, factoryErr = ctl.repoFactory(conf)
		if factoryErr != nil {
			return factoryErr
		}
	}
	return ctl.repo.ResetData(ctx)
}

func (ctl *IndexTestRepoController) Close() error {
	if ctl.repo != nil {
		return ctl.repo.Close()
	}
	return nil
}

func (ctl *IndexTestRepoController) GetIconCount(ctx context.Context) (int, error) {
	return ctl.repo.GetIconCount(ctx)
}

func (ctl *IndexTestRepoController) DescribeIcon(ctx context.Context, iconName string) (domain.IconDescriptor, error) {
	return ctl.repo.DescribeIcon(ctx, iconName)
}

func (ctl *IndexTestRepoController) DescribeAllIcons(ctx context.Context) ([]domain.IconDescriptor, error) {
	return ctl.repo.DescribeAllIcons(ctx)
}

func (ctl *IndexTestRepoController) CreateIcon(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.CreateIcon(ctx, iconName, iconfile, modifiedBy, createSideEffect)
}

func (ctl *IndexTestRepoController) AddIconfileToIcon(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.AddIconfileToIcon(ctx, iconName, iconfile, modifiedBy, createSideEffect)
}

func (ctl *IndexTestRepoController) AddTag(ctx context.Context, iconName string, tag string, modifiedBy string) error {
	return ctl.repo.AddTag(ctx, iconName, tag, modifiedBy)
}

func (ctl *IndexTestRepoController) GetExistingTags(ctx context.Context) ([]string, error) {
	return ctl.repo.GetExistingTags(ctx)
}

func (ctl *IndexTestRepoController) DeleteIcon(ctx context.Context, iconName string, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.DeleteIcon(ctx, iconName, modifiedBy, createSideEffect)
}

func (ctl *IndexTestRepoController) GetIconFileCount(ctx context.Context) (int, error) {
	return ctl.repo.GetIconFileCount(ctx)
}

func (ctl *IndexTestRepoController) GetTagRelationCount(ctx context.Context) (int, error) {
	return ctl.repo.GetTagRelationCount(ctx)
}

func (ctl *IndexTestRepoController) DeleteIconfile(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor, modifiedBy string, createSideEffect func() error) error {
	return ctl.repo.DeleteIconfile(ctx, iconName, iconfile, modifiedBy, createSideEffect)
}

func NewTestPgRepo(conf *config.Options) (TestIndexRepository, error) {
	connection, err := pgdb.NewDBConnection(*conf)
	if err != nil {
		return nil, err
	}
	_, schemaErr := pgdb.OpenSchema(*conf, connection)
	if schemaErr != nil {
		return nil, schemaErr
	}

	sqlDb := pgdb.NewPgRepository(connection)
	return &PgTestRepository{
		PgRepository: &sqlDb,
	}, nil
}

func NewTestDynamodbRepo(conf *config.Options) (TestIndexRepository, error) {
	connection, err := dynamodb.NewDynamodbRepository(conf)
	if err != nil {
		return nil, err
	}

	return &DynamodbTestRepository{connection}, nil
}

type IndexingTestSuite struct {
	*suite.Suite
	config             config.Options
	testRepoController IndexTestRepoController
	ctx                context.Context
}

func (s *IndexingTestSuite) SetupSuite() {
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	s.config = conf
	s.config.DBSchemaName = "itest_repositories"
	s.ctx = logging.Get().WithContext(context.TODO())
}

func (s *IndexingTestSuite) TearDownSuite() {
	s.testRepoController.Close()
}

func (s *IndexingTestSuite) BeforeTest(suiteName, testName string) {
	err := s.testRepoController.ResetRepo(s.ctx, &s.config)
	if err != nil {
		s.FailNow("", "failed to reset test data: %v", err)
	}
}

func (s *IndexingTestSuite) equalIconAttributes(icon1 domain.Icon, icon2 domain.IconDescriptor, expectedTags []string) {
	s.Equal(icon1.Name, icon2.Name)
	s.Equal(icon1.ModifiedBy, icon2.ModifiedBy)
	if expectedTags != nil {
		s.Equal(expectedTags, icon2.Tags)
	}
}

func (s *IndexingTestSuite) getIconCount(ctx context.Context) (int, error) {
	return s.testRepoController.GetIconCount(ctx)
}

var DefaultIndexTestRepoController IndexTestRepoController = IndexTestRepoController{
	repoFactory: func(conf *config.Options) (TestIndexRepository, error) {
		return NewTestPgRepo(conf)
	},
}

var DynamodbIndexTestRepoController IndexTestRepoController = IndexTestRepoController{
	repoFactory: func(conf *config.Options) (TestIndexRepository, error) {
		return NewTestDynamodbRepo(conf)
	},
}

func IndexProvidersToTest() []IndexTestRepoController {
	if len(os.Getenv("PG_ONLY")) > 0 {
		fmt.Print(">>>>>>>>>>> Indexing provider: PG_ONLY\n")
		return []IndexTestRepoController{DefaultIndexTestRepoController}
	}
	if len(os.Getenv("DYNAMODB_ONLY")) > 0 {
		fmt.Print(">>>>>>>>>>> Indexing provider: DYNAMODB_ONLY\n")
		return []IndexTestRepoController{DynamodbIndexTestRepoController}
	}
	fmt.Print(">>>>>>>>>>> Indexing provider: ALL\n")
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
			logging.Get().With().Str("test_sequence_name", "indexing tests").Logger().WithContext(context.TODO()),
		})
	}

	return all
}
