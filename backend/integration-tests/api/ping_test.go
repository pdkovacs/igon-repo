package apitests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/pdkovacs/igo-repo/pkg/server"
	"github.com/stretchr/testify/assert"
)

func setup() int {
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

func teardown() {
	server.ListenerKiller()
}

type Body struct {
	Message string
}

func testPing(t *testing.T, serverPort int) {
	fmt.Printf("Hello, test:%d.....\n", serverPort)
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ping", serverPort), nil)
	if err != nil {
		panic(err)
	}
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	byteBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	var body Body
	err = json.Unmarshal(byteBody, &body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cheers: %s\n", body.Message)
	assert.Equal(t, "pong", body.Message)
	fmt.Printf("Bye, test:%d.....\n", serverPort)
}

func TestPing(t *testing.T) {
	defer teardown()
	serverPort := setup()
	testPing(t, serverPort)
}

func TestPong(t *testing.T) {
	defer teardown()
	serverPort := setup()
	testPing(t, serverPort)
}
