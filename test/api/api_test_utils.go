package api

import (
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pdkovacs/igo-repo/app"
	"github.com/pdkovacs/igo-repo/config"
	httpadapter "github.com/pdkovacs/igo-repo/http"
	"github.com/pdkovacs/igo-repo/repositories"
	"github.com/pdkovacs/igo-repo/test/common"
	repositories_itests "github.com/pdkovacs/igo-repo/test/repositories"
	"github.com/pdkovacs/igo-repo/test/testdata"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type apiTestSuite struct {
	suite.Suite
	defaultConfig config.Options
	server        httpadapter.Server
	testDBRepo    repositories.DatabaseRepository
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

	repositories_itests.DeleteDBData(s.testDBRepo.ConnectionPool)
	s.testDBRepo.Close()
}

// startTestServer starts a test server
func (s *apiTestSuite) startTestServer(conf config.Options) {

	log.SetLevel(log.DebugLevel)

	db, dbErr := repositories.InitDBRepo(conf)
	if dbErr != nil {
		panic(dbErr)
	}
	s.testDBRepo = *db

	git := &repositories.GitRepository{Location: conf.IconDataLocationGit}
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
	s.server = httpadapter.Server{API: httpadapter.API{
		IconService: &app.GetAPI().IconService,
	}}
	go s.server.SetupAndStart(conf, func(port int) {
		s.client.serverPort = port
		log.Infof("Server is listening on port %d", port)
		wg.Done()
	})
	wg.Wait()
}

// terminateTestServer terminates a test server
func (s *apiTestSuite) terminateTestServer() {
	s.server.KillListener()
}
