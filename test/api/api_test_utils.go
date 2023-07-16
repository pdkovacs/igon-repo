package api_tests

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	"iconrepo/internal/app"
	"iconrepo/internal/app/security/authn"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"
	"iconrepo/internal/repositories/blobstore/git"
	"iconrepo/internal/repositories/indexing/pgdb"
	blobstore_tests "iconrepo/test/repositories/blobstore"
	git_tests "iconrepo/test/repositories/blobstore/git"
	"iconrepo/test/repositories/indexing/pg"
	"iconrepo/test/test_commons"
	"iconrepo/test/testdata"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite struct {
	*suite.Suite
	config                  config.Options
	stopServer              func()
	indexRepo               pg.TestIndexRepository
	TestBlobstoreController blobstore_tests.TestBlobstoreController
	Client                  apiTestClient
	logger                  zerolog.Logger
	testSequenceId          string
	xid                     string
}

func apiTestSuites(testSequenceName string, gitProviders []blobstore_tests.TestBlobstoreController) []ApiTestSuite {
	os.Setenv("LOG_LEVEL", "debug")

	all := []ApiTestSuite{}
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	conf.DBSchemaName = testSequenceName
	conf.LocalGitRepo = fmt.Sprintf("%s_%s", conf.LocalGitRepo, testSequenceName)
	for _, repoController := range gitProviders {
		suiteToEmbed := new(suite.Suite)
		all = append(all, ApiTestSuite{
			suiteToEmbed,
			conf,
			nil,
			nil,
			repoController,
			apiTestClient{},
			logging.Get().With().Str("test_sequence_name", testSequenceName).Logger(),
			testSequenceName,
			"",
		})
	}
	return all
}

func (s *ApiTestSuite) SetupSuite() {
	if s.config.DBSchemaName == "" {
		s.FailNow("%v", "No config set by the suite extender")
	}
	s.config.LogLevel = logging.DebugLevel

	// testDBConn and testDBREpo will be only used to read for verification
	testDBConn, testDBErr := pgdb.NewDBConnection(s.config)
	if testDBErr != nil {
		s.FailNow("%v", testDBErr)
	}
	testDbRepo := pgdb.NewDBRepository(testDBConn)
	s.indexRepo = pg.NewTestDbRepositoryFromSQLDB(&testDbRepo)

	var apiTokenErr error
	s.config.GitlabAccessToken, apiTokenErr = git_tests.GitTestGitlabAPIToken()
	if apiTokenErr != nil {
		s.FailNow("%v", apiTokenErr)
	}

	s.config.PasswordCredentials = []config.PasswordCredentials{
		testdata.DefaultCredentials,
	}
	s.config.AuthenticationType = authn.SchemeBasic
	s.config.ServerPort = 0

	s.logger = logging.CreateUnitLogger(s.logger, "apiTestSuite")
}

func (s *ApiTestSuite) TearDownSuite() {
	s.indexRepo.Close()
}

func (s *ApiTestSuite) initTestCaseConfig(testName string) {
	s.xid = xid.New().String()
	s.logger = s.logger.With().Str("app_xid", s.xid).Logger()

	git_tests.SetupGitlabTestCaseConfig(&s.config, s.testSequenceId, s.xid)

	createRepoErr := s.TestBlobstoreController.ResetRepository(&s.config)
	if createRepoErr != nil {
		s.FailNow("%v", createRepoErr)
	}

	err := s.indexRepo.ResetData()
	if err != nil {
		s.FailNow("%v", err)
	}
}

func (s *ApiTestSuite) BeforeTest(suiteName string, testName string) {
	s.initTestCaseConfig(testName)
	s.config.EnableBackdoors = true
	startErr := s.startApp(s.config)
	if startErr != nil {
		s.FailNow("%v", startErr)
	}
}

func (s *ApiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	os.Unsetenv(git.SimulateGitCommitFailureEnvvarName)
	deleteRepoErr := s.TestBlobstoreController.DeleteRepository()
	if deleteRepoErr != nil {
		s.logger.Error().Err(deleteRepoErr).Str("project", s.TestBlobstoreController.String()).Msg("failed to delete testGitRepo")
	}
}

func (s *ApiTestSuite) startApp(serverConfig config.Options) error {
	var wg sync.WaitGroup
	wg.Add(1)
	var startFailure error
	go func() {
		startErr := app.Start(serverConfig, func(port int, stopServer func()) {
			s.Client.serverPort = port
			s.stopServer = stopServer
			wg.Done()
		})
		startFailure = startErr
		if startErr != nil {
			wg.Done()
		}
	}()
	wg.Wait()
	if startFailure != nil {
		return fmt.Errorf("failed to start server: %w", startFailure)
	}
	return nil
}

// terminateTestServer terminates a test server
func (s *ApiTestSuite) terminateTestServer() {
	fmt.Fprintln(os.Stderr, "Stopping test server...")
	if s != nil && s.stopServer != nil {
		s.stopServer()
	}
}
