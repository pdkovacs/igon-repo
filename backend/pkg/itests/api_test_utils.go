package itests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/pdkovacs/igo-repo/backend/pkg/server"
)

var serverPort int

// startTestServer starts a test server
func startTestServer() {
	var wg sync.WaitGroup
	wg.Add(1)
	go server.SetupAndStart(0, func(port int) {
		serverPort = port
		wg.Done()
	})
	wg.Wait()
}

// terminateTestServer terminates a test server
func terminateTestServer() {
	server.KillListener()
}

func get(requestPath string, respJSON interface{}) error {
	request, requestCreationError := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d%s", serverPort, requestPath), nil)
	if requestCreationError != nil {
		return fmt.Errorf("Failed to create request: %w", requestCreationError)
	}
	client := http.Client{}
	response, requestExecutionError := client.Do(request)
	if requestExecutionError != nil {
		return fmt.Errorf("Failed to execute request: %w", requestExecutionError)
	}
	byteBody, responseReadError := ioutil.ReadAll(response.Body)
	if responseReadError != nil {
		return fmt.Errorf("Failed to read response body: %w", responseReadError)
	}
	jsonUnmarshalError := json.Unmarshal(byteBody, respJSON)
	if jsonUnmarshalError != nil {
		return fmt.Errorf("Failed to unmarshal JSON response: %w", jsonUnmarshalError)
	}
	return nil
}

// defaultAuth holds the test PasswordCredentials
var defaultAuth = passwordCredentials{user: "ux", password: "ux"}
