package api

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	app "igo-repo/internal/app"
	"igo-repo/internal/config"
	httpadapter "igo-repo/internal/http"
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
	server        httpadapter.Stoppable
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
	repoConn, testDBErr := repositories.NewDBConnection(s.defaultConfig, logging.CreateUnitLogger(rootAPILogger, "test-db-connection"))
	if testDBErr != nil {
		panic(testDBErr)
	}
	s.testDBRepo = *repositories.NewDBRepository(repoConn, logging.CreateUnitLogger(rootAPILogger, "test-db-repository"))
	s.testGitRepo = repositories_itests.GitTestRepo{
		GitRepository: *repositories.NewGitRepository(s.defaultConfig.IconDataLocationGit, logging.CreateUnitLogger(rootAPILogger, "test-git-repository")),
	}
}

func (s *apiTestSuite) BeforeTest(suiteName string, testName string) {
	serverConfig := common.CloneConfig(s.defaultConfig)
	serverConfig.EnableBackdoors = true
	s.startApp(serverConfig)
}

func (s *apiTestSuite) startApp(serverConfig config.Options) {
	var wg sync.WaitGroup
	wg.Add(1)
	go app.Start(serverConfig, func(port int, server httpadapter.Stoppable) {
		s.client.serverPort = port
		s.server = server
		wg.Done()
	})
	wg.Wait()
}

func (s *apiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	err := repositories_itests.DeleteTestGitRepo(s.defaultConfig.IconDataLocationGit)
	if err != nil {
		panic(err)
	}
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)

	repositories_itests.DeleteDBData(s.testDBRepo.Conn.Pool)
}

// terminateTestServer terminates a test server
func (s *apiTestSuite) terminateTestServer() {
	fmt.Fprintln(os.Stderr, "Stopping test server...")
	s.server.Stop()
}
