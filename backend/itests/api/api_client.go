package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
)

var errUnexpectedStatusCode = errors.New("unexpected status code")

var authenticationBackdoorPath = "/backdoor/authentication"

type apiTestClient struct {
	serverPort int
}

type requestCredentials struct {
	headerName  string
	headerValue string
}

type requestType struct {
	path               string
	credentials        *requestCredentials
	jar                http.CookieJar
	expectedStatusCode int
	json               bool
	body               interface{}
	respBodyProto      interface{}
}

type responseType struct {
	headers map[string][]string
	body    interface{}
}

func (s *apiTestClient) get(req *requestType) (responseType, error) {
	return s.req("GET", req)
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

func (s *apiTestClient) req(method string, req *requestType) (responseType, error) {
	body, errBodyEncode := encodeRequestBody(req.body, req.json)
	if errBodyEncode != nil {
		return responseType{}, errBodyEncode
	}

	request, requestCreationError := http.NewRequest(
		method,
		fmt.Sprintf("http://localhost:%d%s", s.serverPort, req.path),
		body,
	)
	if requestCreationError != nil {
		return responseType{}, fmt.Errorf("Failed to create request: %w", requestCreationError)
	}

	var credentials requestCredentials
	if req.credentials == nil {
		if req.jar == nil {
			var credError error
			credentials, credError = makeRequestCredentials(auxiliaries.BasicAuthentication, defaultCredentials.User, defaultCredentials.Password)
			if credError != nil {
				return responseType{}, fmt.Errorf("Failed to create default request credentials: %w", credError)
			}
		}
	} else {
		credentials = *req.credentials
	}
	if credentials.headerName != "" {
		request.Header.Set(credentials.headerName, credentials.headerValue)
	}

	if req.json {
		request.Header.Set("Content-Type", "application/json")
	}
	client := http.Client{
		Jar: req.jar,
	}
	resp, requestExecutionError := client.Do(request)
	if requestExecutionError != nil {
		return responseType{}, fmt.Errorf("Failed to execute request: %w", requestExecutionError)
	}
	if resp.StatusCode != req.expectedStatusCode {
		return responseType{}, fmt.Errorf("%w expected: %d, got: %d", errUnexpectedStatusCode, req.expectedStatusCode, resp.StatusCode)
	}

	if req.respBodyProto != nil {
		byteBody, responseReadError := io.ReadAll(resp.Body)
		if responseReadError != nil {
			return responseType{}, fmt.Errorf("Failed to read response body: %w", responseReadError)
		}
		jsonUnmarshalError := json.Unmarshal(byteBody, req.respBodyProto)
		if jsonUnmarshalError != nil {
			return responseType{}, fmt.Errorf("Failed to unmarshal JSON response: %w", jsonUnmarshalError)
		}
	}

	request.URL.Path = "/"
	if req.jar != nil {
		req.jar.SetCookies(request.URL, resp.Cookies())
	}

	var responseBody = req.respBodyProto
	req.respBodyProto = nil
	return responseType{
		headers: resp.Header,
		body:    responseBody,
	}, nil
}
