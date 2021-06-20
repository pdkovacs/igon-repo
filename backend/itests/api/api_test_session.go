package api

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/security/authr"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
)

type apiTestSession struct {
	apiTestClient
	cjar http.CookieJar
}

func (client *apiTestClient) login(credentials *requestCredentials) (apiTestSession, error) {
	if credentials == nil {
		calculatedCredentials, credError := makeRequestCredentials(auxiliaries.BasicAuthentication, defaultCredentials.Username, defaultCredentials.Password)
		if credError != nil {
			return apiTestSession{}, fmt.Errorf("Failed to create default request credentials: %w", credError)
		}
		credentials = &calculatedCredentials
	}
	cjar := client.MustCreateCookieJar()

	resp, postError := client.post(&testRequest{
		path:        "/login",
		credentials: credentials,
		jar:         cjar,
		json:        true,
		body:        credentials,
	})
	if postError != nil {
		return apiTestSession{}, fmt.Errorf("failed to login: %w", postError)
	}
	if resp.statusCode != 200 {
		return apiTestSession{}, fmt.Errorf(
			"failed to login with status code %d: %w",
			resp.statusCode,
			errors.New("authentication error"),
		)
	}

	return apiTestSession{
		apiTestClient: apiTestClient{
			serverPort: client.serverPort,
		},
		cjar: cjar,
	}, nil
}

func (client *apiTestClient) mustLogin(credentials *requestCredentials) apiTestSession {
	session, err := client.login(credentials)
	if err != nil {
		panic(err)
	}
	return session
}

func (session *apiTestSession) get(request *testRequest) (testResponse, error) {
	request.jar = session.cjar
	return session.sendRequest("GET", request)
}

func (session *apiTestSession) post(request *testRequest) (testResponse, error) {
	request.jar = session.cjar
	return session.sendRequest("POST", request)
}

func (session *apiTestSession) put(request *testRequest) (testResponse, error) {
	request.jar = session.cjar
	return session.sendRequest("PUT", request)
}

func (session *apiTestSession) setAuthorization(requestedAuthorization web.BackdoorAuthorization) (testResponse, error) {
	var err error
	var resp testResponse
	credentials := session.makeRequestCredentials(defaultCredentials)
	if err != nil {
		panic(err)
	}
	resp, err = session.sendRequest("PUT", &testRequest{
		path:        authenticationBackdoorPath,
		credentials: &credentials,
		jar:         session.cjar,
		json:        true,
		body:        requestedAuthorization,
	})
	return resp, err
}

func (session *apiTestSession) mustSetAuthorization(requestedPermissions []authr.PermissionID) {
	requestedAuthorization := web.BackdoorAuthorization{
		Username:    defaultCredentials.Username,
		Permissions: requestedPermissions,
	}
	resp, err := session.setAuthorization(requestedAuthorization)
	if err != nil {
		panic(err)
	}
	if resp.statusCode != 200 {
		panic(fmt.Sprintf("Unexpected status code: %d", resp.statusCode))
	}
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func (session *apiTestSession) createIcon(iconName string, initialIconfile []byte) (int, domain.Iconfile, error) {
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

	creds := session.makeRequestCredentials(defaultCredentials)

	headers := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	resp, err = session.sendRequest("POST", &testRequest{
		path:          "/icon",
		credentials:   &creds,
		jar:           session.cjar,
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
