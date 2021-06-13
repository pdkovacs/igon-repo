package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
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

func (c *apiTestClient) get(req *testRequest) (testResponse, error) {
	return c.doRequest("GET", req)
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

func (c *apiTestClient) doRequest(method string, req *testRequest) (testResponse, error) {
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

	var credentials requestCredentials
	if req.credentials == nil {
		if req.jar == nil {
			var credError error
			credentials, credError = makeRequestCredentials(auxiliaries.BasicAuthentication, defaultCredentials.Username, defaultCredentials.Password)
			if credError != nil {
				return testResponse{}, fmt.Errorf("Failed to create default request credentials: %w", credError)
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

func (c *apiTestClient) setAuthorization(cjar http.CookieJar, requestedAuthorization web.BackdoorAuthorization) (testResponse, error) {
	var err error
	var resp testResponse
	credentials := c.makeRequestCredentials(defaultCredentials)
	if err != nil {
		panic(err)
	}
	resp, err = c.doRequest("PUT", &testRequest{
		path:        authenticationBackdoorPath,
		credentials: &credentials,
		jar:         cjar,
		json:        true,
		body:        requestedAuthorization,
	})
	return resp, err
}

func (c *apiTestClient) mustSetAuthorization(cjar http.CookieJar, requestedPermissions []authr.PermissionID) {
	requestedAuthorization := web.BackdoorAuthorization{
		Username:    defaultCredentials.Username,
		Permissions: requestedPermissions,
	}
	resp, err := c.setAuthorization(cjar, requestedAuthorization)
	if err != nil {
		panic(err)
	}
	if resp.statusCode != 200 {
		panic(fmt.Sprintf("Unexpected status code: %d", resp.statusCode))
	}
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func (c *apiTestClient) createIcon(cjar http.CookieJar, iconName string, initialIconfile []byte) (int, domain.Iconfile, error) {
	var err error
	var resp testResponse

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	var fw io.Writer
	if fw, err = w.CreateFormField("iconName"); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, strings.NewReader(iconName)); err != nil {
		panic(err)
	}

	if fw, err = w.CreateFormFile("iconfile", iconName); err != nil {
		panic(err)
	}
	if _, err = io.Copy(fw, bytes.NewReader([]byte(base64.StdEncoding.EncodeToString(initialIconfile)))); err != nil {
		panic(err)
	}
	w.Close()

	creds := c.makeRequestCredentials(defaultCredentials)

	headers := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	resp, err = c.doRequest("POST", &testRequest{
		path:          "/icon",
		credentials:   &creds,
		jar:           cjar,
		headers:       headers,
		body:          b.Bytes(),
		respBodyProto: &domain.Iconfile{},
	})
	if err != nil {
		return resp.statusCode, domain.Iconfile{}, err
	}

	if respIconfile, ok := resp.body.(*domain.Iconfile); ok {
		return resp.statusCode, *respIconfile, nil
	}

	return resp.statusCode, domain.Iconfile{}, errors.New(fmt.Sprintf("failed to cast %T to domain.Iconfile", resp.body))
}
