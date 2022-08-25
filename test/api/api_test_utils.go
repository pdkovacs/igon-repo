package api

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	app "igo-repo/internal/app"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"igo-repo/internal/repositories"
	"igo-repo/test/common"
	repositories_itests "igo-repo/test/repositories"
	"igo-repo/test/testdata"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/suite"
)

var rootAPILogger = logging.CreateRootLogger(logging.DebugLevel)

type apiTestSuite struct {
	suite.Suite
	defaultConfig config.Options
	stopServer    func()
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

	// testDBConn and testDBREpo will be only used to read for verification
	testDBConn, testDBErr := repositories.NewDBConnection(s.defaultConfig, logging.CreateUnitLogger(rootAPILogger, "test-db-connection"))
	if testDBErr != nil {
		panic(testDBErr)
	}
	s.testDBRepo = *repositories.NewDBRepository(testDBConn, logging.CreateUnitLogger(rootAPILogger, "test-db-repository"))

	s.testGitRepo = *repositories_itests.NewGitTestRepo(serverConfig.IconDataLocationGit, rootAPILogger)
	repositories_itests.DeleteDBData(s.testDBRepo.Conn.Pool)
	serverConfig.EnableBackdoors = true
	s.startApp(serverConfig)
}

func (s *apiTestSuite) startApp(serverConfig config.Options) {
	var wg sync.WaitGroup
	wg.Add(1)
	go app.Start(serverConfig, func(port int, stopServer func()) {
		s.client.serverPort = port
		s.stopServer = stopServer
		wg.Done()
	})
	wg.Wait()
}

func (s *apiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)
}

// terminateTestServer terminates a test server
func (s *apiTestSuite) terminateTestServer() {
	fmt.Fprintln(os.Stderr, "Stopping test server...")
	s.stopServer()
}
