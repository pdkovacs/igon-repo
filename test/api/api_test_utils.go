package api

import (
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pdkovacs/igo-repo/api"
	"github.com/pdkovacs/igo-repo/config"
	"github.com/pdkovacs/igo-repo/repositories"
	"github.com/pdkovacs/igo-repo/test/api/testdata"
	"github.com/pdkovacs/igo-repo/test/common"
	repositories_itests "github.com/pdkovacs/igo-repo/test/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type apiTestSuite struct {
	suite.Suite
	defaultConfig config.Options
	server        api.Server
	testGitRepo   repositories_itests.GitTestRepo
	client        apiTestClient
}

func (s *apiTestSuite) SetupSuite() {
	s.defaultConfig = common.CloneConfig(config.GetDefaultConfiguration())
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

	repositories_itests.DeleteDBData(s.server.Repositories.DB.ConnectionPool)
	s.server.Repositories.DB.Close()
}

// startTestServer starts a test server
func (s *apiTestSuite) startTestServer(options config.Options) {
	options.ServerPort = 0
	var wg sync.WaitGroup
	wg.Add(1)
	s.server = api.Server{}
	go s.server.SetupAndStart(options, func(port int) {
		s.client.serverPort = port
		log.Infof("Server is listening on port %d", port)
		wg.Done()
	})
	wg.Wait()
	s.testGitRepo.GitRepository = *s.server.Repositories.Git
}

// terminateTestServer terminates a test server
func (s *apiTestSuite) terminateTestServer() {
	s.server.KillListener()
}
