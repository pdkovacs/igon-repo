package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	repositories_itests "github.com/pdkovacs/igo-repo/backend/itests/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type apiTestSuite struct {
	suite.Suite
	defaultOptions auxiliaries.Options
	server         web.Server
	serverPort     int
}

func manageTestResourcesAfterEach() {
}

func (s *apiTestSuite) SetupSuite() {
	s.defaultOptions = auxiliaries.GetDefaultConfiguration()
	s.defaultOptions.PasswordCredentials = []auxiliaries.PasswordCredentials{
		{User: "ux", Password: "ux"},
	}
	s.defaultOptions.ServerPort = 0

	s.defaultOptions.DBSchemaName = "itest_api"
	s.defaultOptions.IconDataCreateNew = "itest-api"
}

func (s *apiTestSuite) BeforeTest(suiteName, testName string) {
	s.startTestServer(s.defaultOptions)
}

func (s *apiTestSuite) AfterTest(suiteName, testName string) {
	s.terminateTestServer()
	err := repositories_itests.DeleteTestGitRepo(s.defaultOptions.IconDataLocationGit)
	if err != nil {
		panic(err)
	}
	os.Unsetenv(repositories.IntrusiveGitTestEnvvarName)

	repositories_itests.DeleteDBData(s.server.DBRepository.ConnectionPool)
	s.server.DBRepository.Close()
}

// startTestServer starts a test server
func (s *apiTestSuite) startTestServer(options auxiliaries.Options) {
	options.ServerPort = 0
	var wg sync.WaitGroup
	wg.Add(1)
	s.server = web.Server{}
	go s.server.SetupAndStart(options, func(port int) {
		s.serverPort = port
		log.Infof("Server is listening on port %d", port)
		wg.Done()
	})
	wg.Wait()
}

// terminateTestServer terminates a test server
func (s *apiTestSuite) terminateTestServer() {
	s.server.KillListener()
}

type requestCredentials struct {
	headerName  string
	headerValue string
}

type request struct {
	path               string
	credentials        *requestCredentials
	expectedStatusCode int
	body               interface{}
	testSuite          *suite.Suite
}

type response struct {
	headers map[string][]string
	body    interface{}
}

func (s *apiTestSuite) get(req *request) (response, error) {
	request, requestCreationError := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d%s", s.serverPort, req.path), nil)
	if requestCreationError != nil {
		return response{}, fmt.Errorf("Failed to create request: %w", requestCreationError)
	}

	var credentials requestCredentials
	if req.credentials == nil {
		var credError error
		credentials, credError = makeRequestCredentials(basicAuthScheme, defaultCredentials.User, defaultCredentials.Password)
		if credError != nil {
			return response{}, fmt.Errorf("Failed to create default request credentials: %w", credError)
		}
	} else {
		credentials = *req.credentials
	}
	if credentials.headerName != "" {
		request.Header.Set(credentials.headerName, credentials.headerValue)
	}

	client := http.Client{}
	resp, requestExecutionError := client.Do(request)
	if requestExecutionError != nil {
		return response{}, fmt.Errorf("Failed to execute request: %w", requestExecutionError)
	}
	if resp.StatusCode != req.expectedStatusCode {
		req.testSuite.Require().FailNow("Unexpected status code", "expected: %d, got: %d", req.expectedStatusCode, resp.StatusCode)
		return response{}, nil
	}

	if req.body != nil {
		byteBody, responseReadError := ioutil.ReadAll(resp.Body)
		if responseReadError != nil {
			return response{}, fmt.Errorf("Failed to read response body: %w", responseReadError)
		}
		jsonUnmarshalError := json.Unmarshal(byteBody, req.body)
		if jsonUnmarshalError != nil {
			return response{}, fmt.Errorf("Failed to unmarshal JSON response: %w", jsonUnmarshalError)
		}
	}

	var responseBody = req.body
	req.body = nil
	return response{
		headers: resp.Header,
		body:    responseBody,
	}, nil
}

// defaultCredentials holds the test PasswordCredentials
var defaultCredentials = auxiliaries.PasswordCredentials{User: "ux", Password: "ux"}
