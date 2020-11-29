package itests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type Body struct {
	Message string
}

type integrationTestSuite struct {
	suite.Suite
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, &integrationTestSuite{})
}

func testPing(s *integrationTestSuite, serverPort int) {
	fmt.Printf("Hello, test:%d.....\n", serverPort)
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/ping", serverPort), nil)
	s.NoError(err)
	client := http.Client{}
	response, err := client.Do(req)
	s.NoError(err)
	byteBody, err := ioutil.ReadAll(response.Body)
	s.NoError(err)
	var body Body
	err = json.Unmarshal(byteBody, &body)
	s.NoError(err)

	fmt.Printf("Cheers: %s\n", body.Message)
	s.Equal("pong", body.Message)
	fmt.Printf("Bye, test:%d.....\n", serverPort)
}

func (s *integrationTestSuite) TestPing() {
	defer terminateTestServer()
	serverPort := startTestServer()
	testPing(s, serverPort)
}

func (s *integrationTestSuite) TestPong() {
	defer terminateTestServer()
	serverPort := startTestServer()
	testPing(s, serverPort)
}
