package api

import (
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	"igo-repo/internal/app"
	"igo-repo/internal/config"
	httpadapter "igo-repo/internal/http"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories"
	"igo-repo/test/common"
	repositories_itests "igo-repo/test/repositories"
	"igo-repo/test/testdata"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
)

var rootAPILogger = logging.CreateRootLogger(logging.DebugLevel)

type Closeable interface {
	Close()
}

type apiTestSuite struct {
	suite.Suite
	defaultConfig config.Options
	server        Closeable
	testDBRepo    repositories.DBRepository
	testGitRepo   repositories_itests.GitTestRepo
	client        apiTestClient
}

func (s *apiTestSuite) SetupSuite() {
	s.defaultConfig = common.GetTestConfig()
	s.defaultConfig.PasswordCredentials = []config.PasswordCredentials{
		testdata.DefaultCredentials,
	}
	s.defaultConfig.AuthenticationType = config.BasicAuthentication
	s.defaultConfig.ServerPort = 0

	s.defaultConfig.DBSchemaName = "itest_api"
	s.defaultConfig.IconDataCreateNew = "itest-api"
}

func (s *apiTestSuite) BeforeTest(suiteName string, testName string) {
	serverConfig := common.CloneConfig(s.defaultConfig)
	serverConfig.EnableBackdoors = true
	s.startTestServer(serverConfig)
}

func (s *apiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	err := repositories_itests.DeleteTestGitRepo(s.defaultConfig.IconDataLocationGit)
	if err != nil {
		panic(err)
	}
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)

	repositories_itests.DeleteDBData(s.testDBRepo.Conn.Pool)
	s.testDBRepo.Conn.Pool.Close()
}

// startTestServer starts a test server
func (s *apiTestSuite) startTestServer(conf config.Options) {

	connection, connOpenErr := repositories.NewDBConnection(conf, log.With().Str("unit", "test-db-connection").Logger())
	if connOpenErr != nil {
		panic(connOpenErr)
	}

	db := repositories.NewDBRepository(connection, log.With().Str("unit", "test-db-repo").Logger())
	s.testDBRepo = *db

	git := repositories.NewGitRepository(conf.IconDataLocationGit, logging.CreateUnitLogger(rootAPILogger, "git-repository"))
	gitErr := git.InitMaybe()
	if gitErr != nil {
		panic(gitErr)
	}
	s.testGitRepo = repositories_itests.GitTestRepo{GitRepository: *git}

	combinedRepo := repositories.RepoCombo{DB: db, Git: git}

	app := app.App{Repository: &combinedRepo}

	conf.ServerPort = 0
	var wg sync.WaitGroup
	wg.Add(1)
	server := httpadapter.CreateServer(
		conf,
		httpadapter.CreateAPI(app.GetAPI(logging.CreateUnitLogger(rootAPILogger, "api")).IconService),
		logging.CreateUnitLogger(rootAPILogger, "server"),
	)
	s.server = &server
	go server.SetupAndStart(conf, func(port int) {
		s.client.serverPort = port
		log.Logger.Info().Msgf("Server is listening on port %d", port)
		wg.Done()
	})
	wg.Wait()
}

// terminateTestServer terminates a test server
func (s *apiTestSuite) terminateTestServer() {
	s.server.Close()
}
