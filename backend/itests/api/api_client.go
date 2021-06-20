package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
)

var authenticationBackdoorPath = "/backdoor/authentication"

type apiTestClient struct {
	serverPort int
}

type requestCredentials struct {
	headerName  string
	headerValue string
}

type testRequest struct {
	path          string
	credentials   *requestCredentials
	jar           http.CookieJar
	headers       map[string]string
	json          bool
	body          interface{}
	respBodyProto interface{}
}

type testResponse struct {
	headers    map[string][]string
	statusCode int
	body       interface{}
}

func (c *apiTestClient) makeRequestCredentials(pwCreds auxiliaries.PasswordCredentials) requestCredentials {
	var username, password string
	if len(pwCreds.Username) == 0 {
		username = defaultCredentials.Username
		password = defaultCredentials.Password
	} else {
		username = pwCreds.Username
		password = pwCreds.Password
	}

	reqCr, err := makeRequestCredentials(auxiliaries.BasicAuthentication, username, password)
	if err != nil {
		panic(err)
	}

	return reqCr
}

func encodeRequestBody(requestBody interface{}, isJSON bool) (io.Reader, error) {
	if requestBody == nil {
		return nil, nil
	}
	if isJSON {
		bodyAsJSON, errUnmarshal := json.Marshal(requestBody)
		if errUnmarshal != nil {
			return nil, fmt.Errorf("failed to unmarshal request body: %w", errUnmarshal)
		}
		return bytes.NewBuffer(bodyAsJSON), nil
	}
	bodyAsBytes, ok := requestBody.([]byte)
	if !ok {
		return nil, fmt.Errorf("I don't know how to convert request body to bytes: %v of %T", requestBody, requestBody)
	}
	return bytes.NewBuffer(bodyAsBytes), nil
}

func (c *apiTestClient) sendRequest(method string, req *testRequest) (testResponse, error) {
	body, errBodyEncode := encodeRequestBody(req.body, req.json)
	if errBodyEncode != nil {
		return testResponse{}, errBodyEncode
	}

	request, requestCreationError := http.NewRequest(
		method,
		fmt.Sprintf("http://localhost:%d%s", c.serverPort, req.path),
		body,
	)
	if requestCreationError != nil {
		return testResponse{}, fmt.Errorf("Failed to create request: %w", requestCreationError)
	}

	if req.credentials != nil && req.credentials.headerName != "" {
		var credentials requestCredentials = *req.credentials
		request.Header.Set(credentials.headerName, credentials.headerValue)
	}

	if req.json {
		request.Header.Set("Content-Type", "application/json")
	}

	for headerName, headerValue := range req.headers {
		request.Header.Set(headerName, headerValue)
	}

	client := http.Client{
		Jar: req.jar,
	}
	resp, requestExecutionError := client.Do(request)
	if requestExecutionError != nil {
		return testResponse{}, fmt.Errorf("Failed to execute request: %w", requestExecutionError)
	}

	if req.respBodyProto != nil {
		byteBody, responseReadError := io.ReadAll(resp.Body)
		if responseReadError != nil {
			return testResponse{}, fmt.Errorf("Failed to read response body: %w", responseReadError)
		}
		jsonUnmarshalError := json.Unmarshal(byteBody, req.respBodyProto)
		if jsonUnmarshalError != nil {
			return testResponse{
				headers:    resp.Header,
				statusCode: resp.StatusCode,
			}, fmt.Errorf("Failed to unmarshal JSON response: %w", jsonUnmarshalError)
		}
	}

	request.URL.Path = "/"
	if req.jar != nil {
		req.jar.SetCookies(request.URL, resp.Cookies())
	}

	var responseBody = req.respBodyProto
	req.respBodyProto = nil
	return testResponse{
		headers:    resp.Header,
		statusCode: resp.StatusCode,
		body:       responseBody,
	}, nil
}

func (c *apiTestClient) MustCreateCookieJar() http.CookieJar {
	cjar, errCreatingJar := cookiejar.New(nil)
	if errCreatingJar != nil {
		panic(errCreatingJar)
	}
	return cjar
}

func (c *apiTestClient) get(req *testRequest) (testResponse, error) {
	return c.sendRequest("GET", req)
}

func (c *apiTestClient) post(req *testRequest) (testResponse, error) {
	return c.sendRequest("POST", req)
}
