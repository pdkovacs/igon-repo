package itests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/server"
	"github.com/stretchr/testify/suite"
)

var defaultOptions auxiliaries.Options

func init() {
	var err error
	defaultOptions, err = auxiliaries.ReadConfiguration("", []string{})
	if err != nil {
		panic(err)
	}
	defaultOptions.PasswordCredentials = []auxiliaries.PasswordCredentials{
		{User: "ux", Password: "ux"},
	}
}

var serverPort int

// startTestServer starts a test server
func startTestServer(options auxiliaries.Options) {
	var wg sync.WaitGroup
	wg.Add(1)
	go server.SetupAndStart(0, options, func(port int) {
		serverPort = port
		wg.Done()
	})
	wg.Wait()
}

// terminateTestServer terminates a test server
func terminateTestServer() {
	server.KillListener()
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

func get(req *request) (response, error) {
	request, requestCreationError := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d%s", serverPort, req.path), nil)
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
