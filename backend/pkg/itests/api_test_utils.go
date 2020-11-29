package itests

import (
	"sync"

	"github.com/pdkovacs/igo-repo/backend/pkg/server"
)

// startTestServer starts a test server
func startTestServer() int {
	var serverPort int
	var wg sync.WaitGroup
	wg.Add(1)
	go server.SetupAndStartServer(func(port int) {
		serverPort = port
		wg.Done()
	})
	wg.Wait()
	return serverPort
}

// terminateTestServer terminates a test server
func terminateTestServer() {
	server.ListenerKiller()
}

// defaultAuth holds the test PasswordCredentials
var defaultAuth = passwordCredentials{user: "ux", password: "ux"}
