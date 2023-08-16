package server

import (
	"context"
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
	blobstore_tests "iconrepo/test/repositories/blobstore"
	git_tests "iconrepo/test/repositories/blobstore/git"
	"iconrepo/test/repositories/indexing"
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
	indexingController      indexing.IndexTestRepoController
	TestBlobstoreController blobstore_tests.TestBlobstoreController
	Client                  apiTestClient
	Ctx                     context.Context
	testSequenceId          string
	xid                     string
}

func apiTestSuites(
	testSequenceName string,
	blobstoreProviders []blobstore_tests.TestBlobstoreController,
	indexingProviders []indexing.IndexTestRepoController,
) []ApiTestSuite {
	os.Setenv("LOG_LEVEL", "debug")

	all := []ApiTestSuite{}
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	conf.DBSchemaName = testSequenceName
	conf.LocalGitRepo = fmt.Sprintf("%s_%s", conf.LocalGitRepo, testSequenceName)
	for _, repoController := range blobstoreProviders {
		for _, indexinController := range indexingProviders {
			suiteToEmbed := new(suite.Suite)
			all = append(all, ApiTestSuite{
				suiteToEmbed,
				conf,
				nil,
				indexinController,
				repoController,
				apiTestClient{},
				logging.Get().With().Str("test_sequence_name", testSequenceName).Logger().WithContext(context.TODO()),
				testSequenceName,
				"",
			})
		}
	}
	return all
}

func (s *ApiTestSuite) SetupSuite() {
	if s.config.DBSchemaName == "" {
		s.FailNow("", "%v", "No config set by the suite extender")
	}
	s.config.LogLevel = logging.DebugLevel

	var apiTokenErr error
	s.config.GitlabAccessToken, apiTokenErr = git_tests.GitTestGitlabAPIToken()
	if apiTokenErr != nil {
		s.FailNow("", "%v", apiTokenErr)
	}

	s.config.PasswordCredentials = []config.PasswordCredentials{
		testdata.DefaultCredentials,
	}
	s.config.AuthenticationType = authn.SchemeBasic
	s.config.ServerPort = 0

	s.Ctx = logging.CreateUnitLogger(*zerolog.Ctx(s.Ctx), "apiTestSuite").WithContext(s.Ctx)
}

func (s *ApiTestSuite) TearDownSuite() {
	s.indexingController.Close()
}

func (s *ApiTestSuite) initTestCaseConfig(testName string) {
	s.xid = xid.New().String()
	s.Ctx = zerolog.Ctx(s.Ctx).With().Str("app_xid", s.xid).Logger().WithContext(s.Ctx)

	git_tests.SetupGitlabTestCaseConfig(&s.config, s.testSequenceId, s.xid)

	createRepoErr := s.TestBlobstoreController.ResetRepository(&s.config)
	if createRepoErr != nil {
		s.FailNow("", "%v", createRepoErr)
	}

	err := s.indexingController.ResetRepo(s.Ctx, &s.config)
	if err != nil {
		s.FailNow("", "%v", err)
	}
}

func (s *ApiTestSuite) BeforeTest(suiteName string, testName string) {
	s.initTestCaseConfig(testName)
	s.config.EnableBackdoors = true
	startErr := s.startApp(s.config)
	if startErr != nil {
		s.FailNow("", "%v", startErr)
	}
}

func (s *ApiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	os.Unsetenv(git.SimulateGitCommitFailureEnvvarName)
	deleteRepoErr := s.TestBlobstoreController.DeleteRepository()
	if deleteRepoErr != nil {
		zerolog.Ctx(s.Ctx).Error().Err(deleteRepoErr).Str("project", s.TestBlobstoreController.String()).Msg("failed to delete testGitRepo")
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
