package api_tests

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	app "igo-repo/internal/app"
	"igo-repo/internal/app/security/authn"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories/gitrepo"
	"igo-repo/internal/repositories/icondb"
	"igo-repo/test/repositories/db_tests"
	"igo-repo/test/repositories/git_tests"
	"igo-repo/test/test_commons"
	"igo-repo/test/testdata"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

var rootAPILogger = logging.CreateRootLogger(logging.DebugLevel)

type ApiTestSuite struct {
	suite.Suite
	config          config.Options
	stopServer      func()
	testDBRepo      icondb.Repository
	TestGitRepo     git_tests.GitTestRepo
	Client          apiTestClient
	logger          zerolog.Logger
	testSequenceId  string
	testCaseCounter int
}

func apiTestSuites(testSequenceId string, gitProviders []git_tests.GitTestRepo) []ApiTestSuite {
	all := []ApiTestSuite{}
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	conf.DBSchemaName = testSequenceId
	conf.LocalGitRepo = fmt.Sprintf("%s_%s", conf.LocalGitRepo, testSequenceId)
	for _, repo := range gitProviders {
		all = append(all, ApiTestSuite{config: conf, TestGitRepo: repo, testSequenceId: testSequenceId})
	}
	return all
}

func (s *ApiTestSuite) SetupSuite() {
	if s.config.DBSchemaName == "" {
		panic("No config set by the suite extender")
	}
	s.config.LogLevel = logging.DebugLevel

	// testDBConn and testDBREpo will be only used to read for verification
	testDBConn, testDBErr := icondb.NewDBConnection(s.config, logging.CreateUnitLogger(rootAPILogger, "test-db-connection"))
	if testDBErr != nil {
		panic(testDBErr)
	}
	s.testDBRepo = icondb.NewDBRepository(testDBConn, logging.CreateUnitLogger(rootAPILogger, "test-db-repository"))

	s.config.GitlabAccessToken = git_tests.GitTestGitlabAPIToken()

	s.config.PasswordCredentials = []config.PasswordCredentials{
		testdata.DefaultCredentials,
	}
	s.config.AuthenticationType = authn.SchemeBasic
	s.config.ServerPort = 0

	s.logger = logging.CreateUnitLogger(rootAPILogger, "apiTestSuite")
}

func (s *ApiTestSuite) initConfig() config.Options {
	s.testCaseCounter++
	conf := test_commons.CloneConfig(s.config)
	conf.GitlabProjectPath = fmt.Sprintf("%s_%s_%d", conf.GitlabProjectPath, s.testSequenceId, s.testCaseCounter)
	switch s.TestGitRepo.(type) {
	case gitrepo.Local:
		conf.GitlabNamespacePath = "" // to guide the test app on which git provider to use
		s.TestGitRepo = git_tests.NewLocalGitTestRepo(conf)
	case gitrepo.Gitlab:
		conf.GitlabNamespacePath = "testing-with-repositories"
		conf.LocalGitRepo = "" // to guide the test app on which git provider to use
		s.TestGitRepo = git_tests.NewGitlabTestRepoClient(conf)
	case nil:
		s.logger.Info().Msg("No testGitRepo set; using default")
		conf.GitlabNamespacePath = "" // to guide the test app on which git provider to use
		s.TestGitRepo = git_tests.NewLocalGitTestRepo(conf)
	}

	git_tests.MustResetTestGitRepo(s.TestGitRepo)
	db_tests.ResetDBData(s.testDBRepo.Conn.Pool)
	return conf
}

func (s *ApiTestSuite) BeforeTest(suiteName string, testName string) {
	conf := s.initConfig()
	conf.EnableBackdoors = true
	startErr := s.startApp(conf)
	if startErr != nil {
		panic(startErr)
	}
}

func (s *ApiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	os.Unsetenv(gitrepo.SimulateGitCommitFailureEnvvarName)

	deleteRepoErr := s.TestGitRepo.Delete()
	if deleteRepoErr != nil {
		s.logger.Error().Msgf("failed to delete testGitRepo %s: %#v", s.TestGitRepo, deleteRepoErr)
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
