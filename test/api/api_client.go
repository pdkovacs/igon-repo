package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/pdkovacs/igo-repo/app/domain"
	"github.com/pdkovacs/igo-repo/config"
	"github.com/pdkovacs/igo-repo/test/testdata"
	log "github.com/sirupsen/logrus"
)

var errUnexpecteHTTPStatus = errors.New("unexpected HTTP status")
var errJSONUnmarshal = errors.New("failed to unmarshal JSON")

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

func (c *apiTestClient) makeRequestCredentials(pwCreds config.PasswordCredentials) requestCredentials {
	var username, password string
	if len(pwCreds.Username) == 0 {
		username = testdata.DefaultCredentials.Username
		password = testdata.DefaultCredentials.Password
	} else {
		username = pwCreds.Username
		password = pwCreds.Password
	}

	reqCr, err := makeRequestCredentials(config.BasicAuthentication, username, password)
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
		return nil, fmt.Errorf("unable to convert request body to bytes: %v of %T", requestBody, requestBody)
	}
	return bytes.NewBuffer(bodyAsBytes), nil
}

func (c *apiTestClient) sendRequest(method string, req *testRequest) (testResponse, error) {
	logger := log.WithField("prefix", "sendRequest")
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
		return testResponse{}, fmt.Errorf("failed to create request: %w", requestCreationError)
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
		return testResponse{}, fmt.Errorf("failed to execute request: %w", requestExecutionError)
	}

	var responseBody interface{}

	if req.respBodyProto != nil {
		byteBody, responseReadError := io.ReadAll(resp.Body)
		if responseReadError != nil {
			return testResponse{}, fmt.Errorf("failed to read response body: %w", responseReadError)
		}
		if _, ok := req.respBodyProto.([]byte); ok {
			responseBody = byteBody
		} else {
			jsonUnmarshalError := json.Unmarshal(byteBody, req.respBodyProto)
			// TODO: We should somehow better handle unmarshalling failed calls as well...
			//       ... using some standard error JSON for example?
			if jsonUnmarshalError != nil {
				logger.Errorf("failed to unmarshal JSON response: %v\n", jsonUnmarshalError)
				return testResponse{
					headers:    resp.Header,
					statusCode: resp.StatusCode,
				}, fmt.Errorf("failed to unmarshal JSON response \"%s\": %w", string(byteBody), errJSONUnmarshal)
			} else {
				responseBody = req.respBodyProto
			}
		}
	}

	request.URL.Path = "/"
	if req.jar != nil {
		req.jar.SetCookies(request.URL, resp.Cookies())
	}

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

func getFilePath(iconName string, fileDescriptor domain.IconfileDescriptor) string {
	return fmt.Sprintf("/icon/%s/format/%s/size/%s", iconName, fileDescriptor.Format, fileDescriptor.Size)
}
