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
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite struct {
	*suite.Suite
	config          config.Options
	stopServer      func()
	testDBRepo      icondb.Repository
	TestGitRepo     git_tests.GitTestRepo
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
			icondb.Repository{},
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
	testDBConn, testDBErr := icondb.NewDBConnection(s.config)
	if testDBErr != nil {
		panic(testDBErr)
	}
	s.testDBRepo = icondb.NewDBRepository(testDBConn)

	s.config.GitlabAccessToken = git_tests.GitTestGitlabAPIToken()

	s.config.PasswordCredentials = []config.PasswordCredentials{
		testdata.DefaultCredentials,
	}
	s.config.AuthenticationType = authn.SchemeBasic
	s.config.ServerPort = 0

	s.logger = logging.CreateUnitLogger(s.logger, "apiTestSuite")
}

func (s *ApiTestSuite) initTestCaseConfig() config.Options {
	s.testCaseCounter++
	s.xid = xid.New().String()
	s.logger = s.logger.With().Str("app_xid", s.xid).Logger()
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
	conf := s.initTestCaseConfig()
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
		s.logger.Error().Err(deleteRepoErr).Str("project", s.TestGitRepo.String()).Msg("failed to delete testGitRepo")
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
