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
	"iconrepo/test/repositories/blobstore_tests/git_tests"
	"iconrepo/test/repositories/indexing_tests"
	"iconrepo/test/test_commons"
	"iconrepo/test/testdata"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite struct {
	*suite.Suite
	config          config.Options
	stopServer      func()
	indexRepo       indexing_tests.TestIndexRepository
	TestBlobstore   git_tests.GitTestRepo
	Client          apiTestClient
	logger          zerolog.Logger
	testSequenceId  string
	testCaseCounter int
	xid             string
}

func apiTestSuites(testSequenceName string, gitProviders []git_tests.GitTestRepo) []ApiTestSuite {
	all := []ApiTestSuite{}
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	conf.DBSchemaName = testSequenceName
	conf.LocalGitRepo = fmt.Sprintf("%s_%s", conf.LocalGitRepo, testSequenceName)
	for _, repo := range gitProviders {
		suiteToEmbed := new(suite.Suite)
		all = append(all, ApiTestSuite{
			suiteToEmbed,
			conf,
			nil,
			nil,
			repo,
			apiTestClient{},
			logging.Get().With().Str("test_sequence_name", testSequenceName).Logger(),
			testSequenceName,
			0,
			"",
		})
	}
	return all
}

func (s *ApiTestSuite) SetupSuite() {
	if s.config.DBSchemaName == "" {
		panic("No config set by the suite extender")
	}
	s.config.LogLevel = logging.DebugLevel

	// testDBConn and testDBREpo will be only used to read for verification
	testDBConn, testDBErr := pgdb.NewDBConnection(s.config)
	if testDBErr != nil {
		panic(testDBErr)
	}
	testDbRepo := pgdb.NewDBRepository(testDBConn)
	s.indexRepo = indexing_tests.NewTestDbRepositoryFromSQLDB(&testDbRepo)

	s.config.GitlabAccessToken = git_tests.GitTestGitlabAPIToken()

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

func (s *ApiTestSuite) initTestCaseConfig() config.Options {
	s.testCaseCounter++
	s.xid = xid.New().String()
	s.logger = s.logger.With().Str("app_xid", s.xid).Logger()
	conf := test_commons.CloneConfig(s.config)
	conf.GitlabProjectPath = fmt.Sprintf("%s_%s_%d", conf.GitlabProjectPath, s.testSequenceId, s.testCaseCounter)
	switch s.TestBlobstore.(type) {
	case git.Local:
		conf.GitlabNamespacePath = "" // to guide the test app on which git provider to use
		s.TestBlobstore = git_tests.NewLocalGitTestRepo(conf)
	case git.Gitlab:
		conf.GitlabNamespacePath = "testing-with-repositories"
		conf.LocalGitRepo = "" // to guide the test app on which git provider to use
		s.TestBlobstore = git_tests.NewGitlabTestRepoClient(conf)
	case nil:
		s.logger.Info().Msg("No testGitRepo set; using default")
		conf.GitlabNamespacePath = "" // to guide the test app on which git provider to use
		s.TestBlobstore = git_tests.NewLocalGitTestRepo(conf)
	}

	git_tests.MustResetTestGitRepo(s.TestBlobstore)
	err := s.indexRepo.ResetDBData()
	if err != nil {
		panic(err)
	}
	return conf
}

func (s *ApiTestSuite) BeforeTest(suiteName string, testName string) {
	conf := s.initTestCaseConfig()
	conf.EnableBackdoors = true
	startErr := s.startApp(conf)
	if startErr != nil {
		panic(startErr)
	}
}

func (s *ApiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	os.Unsetenv(git.SimulateGitCommitFailureEnvvarName)

	deleteRepoErr := s.TestBlobstore.Delete()
	if deleteRepoErr != nil {
		s.logger.Error().Err(deleteRepoErr).Str("project", s.TestBlobstore.String()).Msg("failed to delete testGitRepo")
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
