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

func (session *apiTestSession) setAuthorization(requestedAuthorization []authr.PermissionID) (testResponse, error) {
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
	resp, err := session.setAuthorization(requestedPermissions)
	if err != nil {
		panic(err)
	}
	if resp.statusCode != 200 {
		panic(fmt.Sprintf("Unexpected status code: %d", resp.statusCode))
	}
}

func (session *apiTestSession) describeAllIcons() ([]web.ResponseIcon, error) {
	resp, err := session.get(&testRequest{
		path:          "/icon",
		jar:           session.cjar,
		respBodyProto: &[]web.ResponseIcon{},
	})
	if err != nil {
		return []web.ResponseIcon{}, fmt.Errorf("GET /icon failed: %w", err)
	}
	icons, ok := resp.body.(*[]web.ResponseIcon)
	if !ok {
		return []web.ResponseIcon{}, fmt.Errorf("failed to cast %T as []domain.Icon", resp.body)
	}
	return *icons, err
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func (session *apiTestSession) createIcon(iconName string, initialIconfile []byte) (int, web.ResponseIcon, error) {
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

	headers := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	resp, err = session.sendRequest("POST", &testRequest{
		path:          "/icon",
		jar:           session.cjar,
		headers:       headers,
		body:          b.Bytes(),
		respBodyProto: &web.ResponseIcon{},
	})
	if err != nil {
		return resp.statusCode, web.ResponseIcon{}, err
	}

	if respIconfile, ok := resp.body.(*web.ResponseIcon); ok {
		return resp.statusCode, *respIconfile, nil
	}

	return resp.statusCode, web.ResponseIcon{}, errors.New(fmt.Sprintf("failed to cast %T to web.IconDTO", resp.body))
}

func (s *apiTestSession) GetIconfile(iconName string, iconfileDescriptor domain.IconfileDescriptor) ([]byte, error) {
	content := []byte{}
	resp, reqErr := s.get(&testRequest{
		path: getFilePath(iconName, iconfileDescriptor),
		body: &content,
	})
	if reqErr != nil {
		return content, fmt.Errorf("failed to retrieve iconfile %v of %s", iconfileDescriptor, iconName)
	}

	if byteArr, ok := resp.body.([]byte); ok {
		return byteArr, nil
	}

	return content, fmt.Errorf("failed to cast the reply %T to []byte while retrieving iconfile %v of %s", resp.body, iconfileDescriptor, iconName)
}

func (session *apiTestSession) addIconfile(iconName string, iconfile domain.Iconfile) (int, domain.Iconfile, error) {
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
	if _, err = io.Copy(fw, bytes.NewReader([]byte(base64.StdEncoding.EncodeToString(iconfile.Content)))); err != nil {
		panic(err)
	}
	w.Close()

	headers := map[string]string{
		"Content-Type": w.FormDataContentType(),
	}

	resp, err = session.sendRequest("POST", &testRequest{
		path:          fmt.Sprintf("/icon%s", iconName),
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

	return resp.statusCode, domain.Iconfile{}, errors.New(fmt.Sprintf("failed to cast %T to domain.Icon", resp.body))
}
