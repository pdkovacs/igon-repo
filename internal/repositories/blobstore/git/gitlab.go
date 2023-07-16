package git

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"iconrepo/internal/app/domain"
	"iconrepo/internal/app/security/authn"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var paths = NewGitFilePaths("")

const gitlabRepoHasAlreadyBeenTaken = "has already been taken"

var transientGitlabRepoCreationErrMessages = []string{
	"The project is still being deleted. Please try again later.",
	gitlabRepoHasAlreadyBeenTaken,
}

func isTransientGitlabRepoCreationErrMessage(message string) bool {
	for _, fragment := range transientGitlabRepoCreationErrMessages {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

type responseFileItem struct {
	FileName        string `json:"file_name"`        // : "key.rb",
	FilePath        string `json:"file_path"`        // : "app/models/key.rb",
	Size            int    `json:"size"`             // : 1476,
	Encoding        string `json:"encoding"`         // : "base64",
	Content         string `json:"content"`          // : "IyA9PSBTY2hlbWEgSW5mb3...",
	ContentSha256   string `json:"content_sha256"`   // : "4c294617b60715c1d218e61164a3abd4808a4284cbc30e6728a01ad9aada4481",
	Ref             string `json:"ref"`              // : "master",
	BlobId          string `json:"blob_id"`          // : "79f7bbd25901e8334750839545a9bd021f0e4c83",
	CommitId        string `json:"commit_id"`        // : "d5a3ff139356ce33e37e73add446f16869741b50",
	LastCommitId    string `json:"last_commit_id"`   // : "570e7b2abdd848b95f2f578043fc23bd6f6fd24d",
	ExecuteFilemode bool   `json:"execute_filemode"` // : false
}

type gitlabProject struct {
	namespacePath string
	path          string
	namespaceId   int
}

func (g gitlabProject) String() string {
	return fmt.Sprintf("%s/%s", g.namespacePath, g.path)
}

type Gitlab struct {
	project    gitlabProject
	mainBranch string
	apikey     string
	logger     zerolog.Logger
	client     http.Client
}

func (repo Gitlab) String() string {
	return fmt.Sprintf("GitLab repository at %s?ref=%s", repo.project, repo.mainBranch)
}

type commitActionType string

const (
	commitActionCreate commitActionType = "create"
	commitActionDelete commitActionType = "delete"
	commitActionMove   commitActionType = "move"
	commitActionUpdate commitActionType = "update"
	commitActionChmod  commitActionType = "chmod"
)

type commitActionOnByteSlice struct {
	Action   commitActionType
	FilePath string
	Content  []byte
}

type commitProperties struct {
	Branch        string         `json:"branch"`
	AuthorName    string         `json:"author_name"`
	CommitMessage string         `json:"commit_message"`
	Actions       []commitAction `json:"actions"`
}

type commitAction struct {
	Action   commitActionType `json:"action"`
	FilePath string           `json:"file_path"`
	Content  *string          `json:"content"`
	Encoding *string          `json:"encoding"`
}

type commitQueryResponseItem struct {
	Id             string `json:"id"`
	CommittedDate  string `json:"committed_date"`
	Message        string `json:"message"`
	AuthorName     string `json:"author_name"`
	AuthorEmail    string `json:"author_email"`
	AuthoredDate   string `json:"authored_date"`
	CommitterName  string `json:"committer_name"`
	CommitterEmail string `json:"committer_email"`
}

type repositoryTreeItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

type projectProperties struct {
	NamespaceId          int    `json:"namespace_id"`
	Path                 string `json:"path"`
	Description          string `json:"description"`
	InitializeWithReadme string `json:"initialize_with_readme"`
}

func NewGitlabRepositoryClient(namespacePath string, projectPath string, branch string, apikey string, logger zerolog.Logger) (Gitlab, error) {
	if len(apikey) == 0 {
		return Gitlab{}, fmt.Errorf("no API token for GitLab repository")
	}
	gitlab := Gitlab{
		project: gitlabProject{
			namespacePath: namespacePath,
			path:          projectPath,
		},
		mainBranch: branch,
		apikey:     apikey,
		client: http.Client{
			Timeout: time.Second * 15,
		},
		logger: logger,
	}

	namespaceId, err := getNamespaceID(gitlab)
	if err != nil {
		return gitlab, err
	}
	gitlab.project.namespaceId = namespaceId

	return gitlab, nil
}

func (g *Gitlab) createCreateProjectBody() (io.Reader, error) {
	projectProps := projectProperties{
		NamespaceId: g.project.namespaceId,
		Path:        g.project.path,
	}
	jsonInBytes, marshalErr := json.Marshal(&projectProps)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal project creation data %#v: %w", projectProps, marshalErr)
	}
	return bytes.NewReader(jsonInBytes), nil
}

func (g *Gitlab) CreateRepository() error {
	sleepBeforeRetryMs := 1000
	maxRetryCount := 20

	retryCount := 0
	for {
		requestBody, requestBodyErr := g.createCreateProjectBody()
		if requestBodyErr != nil {
			return fmt.Errorf("failed to create request body for creating test repository: %w", requestBodyErr)
		}
		statusCode, _, responseBody, err := g.sendRequest("POST", "/projects", requestBody)
		if err != nil || (statusCode != 201 && statusCode != 400) {
			return fmt.Errorf("failed to create project: (%d) %s -- %w", statusCode, responseBody, err)
		}
		if statusCode == 400 && isTransientGitlabRepoCreationErrMessage(responseBody) {
			retryCount++
			if retryCount >= maxRetryCount {
				panic("Too many retries creating GitLab repo")
			}

			requestBodyStr, readRequestBodyErr := io.ReadAll(requestBody)
			if readRequestBodyErr != nil {
				return fmt.Errorf("failed to read create project request body: %w", readRequestBodyErr)
			}
			g.logger.Debug().Err(requestBodyErr).
				Str("request-body", string(requestBodyStr)).
				Str("project", g.project.String()).
				Int("sleep-ms-before-retry", sleepBeforeRetryMs).
				Msg("Transient error while creating repository")
			time.Sleep(time.Duration(sleepBeforeRetryMs) * time.Millisecond)
			if strings.Contains(responseBody, gitlabRepoHasAlreadyBeenTaken) {
				g.DeleteRepository()
				time.Sleep(time.Duration(sleepBeforeRetryMs) * time.Millisecond)
			}
			continue
		}
		g.logger.Info().Str("project", g.project.String()).Msg("GitLab repository created")
		return nil
	}
}

func (g *Gitlab) ResetRepository() error {
	deleteRepoErr := g.DeleteRepository()
	if deleteRepoErr != nil {
		panic(deleteRepoErr)
	}
	return g.CreateRepository()
}

func (g *Gitlab) DeleteRepository() error {
	statusCode, _, body, err := g.sendRequest("DELETE", fmt.Sprintf("/projects/%s", url.PathEscape(g.project.String())), nil)
	if err != nil || (statusCode != 202 && statusCode != 404) {
		return fmt.Errorf("failed to delete gitlab repository: (%d) %s -- %w", statusCode, body, err)
	}
	g.logger.Info().Str("project", g.project.String()).Msg("GitLab repository deleted")
	return nil
}

func (g *Gitlab) GetIconfiles() ([]string, error) {
	statusCode, _, body, err := g.sendRequest("GET", fmt.Sprintf("/projects/%s/repository/tree?ref=%s&recursive=true", url.PathEscape(g.project.String()), g.mainBranch), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to get repository tree from GitLab repo: %w", err)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("failed to get repository tree from GitLab repo (%d) %s -- %w", statusCode, body, err)
	}

	tree := []repositoryTreeItem{}
	jsonErr := json.Unmarshal([]byte(body), &tree)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab repository tree response: %w", jsonErr)
	}

	fileList := []string{}

	for _, treeItem := range tree {
		if treeItem.Type == "blob" {
			fileList = append(fileList, treeItem.Path)
		}
	}

	return fileList, nil
}

func (g *Gitlab) createCommitBody(authorName string, commitMessage string, actionsIn []commitActionOnByteSlice) (io.Reader, error) {
	commActs := make([]commitAction, len(actionsIn))

	for index, actionIn := range actionsIn {
		commActs[index].Content = nil
		if len(actionIn.Content) > 1 {
			encodedContent := base64.StdEncoding.EncodeToString(actionIn.Content)
			commActs[index].Content = &encodedContent
			encType := "base64"
			commActs[index].Encoding = &encType
		}
		commActs[index].Action = actionIn.Action
		commActs[index].FilePath = actionIn.FilePath
	}

	commitProps := commitProperties{
		Branch:        g.mainBranch,
		AuthorName:    authorName,
		CommitMessage: commitMessage,
		Actions:       commActs,
	}
	jsonInBytes, marshalErr := json.Marshal(&commitProps)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal commit data %#v: %w", commitProps, marshalErr)
	}
	return bytes.NewReader(jsonInBytes), nil
}

// GetAbsolutePathToIconfile implements repositories_tests.gitTestRepo
func (Gitlab) GetAbsolutePathToIconfile(string, domain.IconfileDescriptor) string {
	panic("unimplemented")
}

// GetStateID implements repositories_tests.gitTestRepo
func (g *Gitlab) GetStateID() (string, error) {
	statusCode, _, body, err := g.sendRequest(
		"GET",
		fmt.Sprintf(
			"/projects/%s/repository/commits?%s",
			url.PathEscape(g.project.String()),
			url.PathEscape(fmt.Sprintf("ref=%s", g.mainBranch)),
		),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to send request to get commit list from GitLab repo: %w", err)
	}
	if statusCode != 200 {
		return "", fmt.Errorf("failed to get commit list from GitLab repo (%d) %s -- %w", statusCode, body, err)
	}

	metadataListResponse := []commitQueryResponseItem{}
	jsonErr := json.Unmarshal([]byte(body), &metadataListResponse)
	if jsonErr != nil {
		return "", fmt.Errorf("failed to unmarshal GitLab commit list response: %w", jsonErr)
	}

	if len(metadataListResponse) < 1 {
		return "", fmt.Errorf("no commit yet in GitLab repository %s", g.project.String())
	}

	return metadataListResponse[0].Id, nil
}

// CheckStatus always returns true for the GitLab repo, since the GitLab service handles consistency (and returns error if it cannot)
func (Gitlab) CheckStatus() (bool, error) {
	return true, nil
}

// GetVersionFor returns the commit ID of the iconfile specified by the method paramters.
// Return empty string in case the file doesn't exist in the repository
func (g *Gitlab) GetVersionFor(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error) {
	commitIdHeaderKey := "X-Gitlab-Commit-Id"

	filePath := paths.getPathComponents(iconName, iconfileDesc).pathToIconfile
	statusCode, header, body, err := g.sendRequest(
		"HEAD",
		fmt.Sprintf(
			"/projects/%s/repository/files/%s?%s",
			url.PathEscape(g.project.String()),
			url.PathEscape(filePath),
			url.PathEscape("ref="+g.mainBranch),
		),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get iconfile commit ID from GitLab repo %s::%s: (%d) %s -- %w", iconName, iconfileDesc.String(), statusCode, body, err)
	}
	if statusCode == 404 {
		return "", nil
	}
	if statusCode != 200 {
		return "", fmt.Errorf("failed to get iconfile commit ID from GitLab repo %s::%s: (%d) %s -- %w", iconName, iconfileDesc.String(), statusCode, body, err)
	}
	return header.Get(commitIdHeaderKey), nil
}

func (g *Gitlab) GetVersionMetadata(commitId string) (CommitMetadata, error) {
	commitMetadata := CommitMetadata{}

	statusCode, _, body, err := g.sendRequest("GET", fmt.Sprintf("/projects/%s/repository/commits/%s", url.PathEscape(g.project.String()), commitId), nil)
	if err != nil {
		return commitMetadata, fmt.Errorf("failed to send request to get commit meta-data for %s from GitLab repo: %w", commitId, err)
	}
	if statusCode != 200 {
		return commitMetadata, fmt.Errorf("failed to get commit meta-data for %s from GitLab repo (%d) %s -- %w", commitId, statusCode, body, err)
	}

	metadataResponse := commitQueryResponseItem{}
	jsonErr := json.Unmarshal([]byte(body), &metadataResponse)
	if jsonErr != nil {
		return commitMetadata, fmt.Errorf("failed to unmarshal GitLab commit meta-data response for %s: %w", commitId, jsonErr)
	}

	commitMetadata, conversionErr := gitlabCommitResponseToMetadata(metadataResponse)
	if conversionErr != nil {
		return commitMetadata, fmt.Errorf("failed to parse commitQueryResponseItem for GitLab commit %s: %w", commitId, conversionErr)
	}

	return commitMetadata, nil
}

func (g *Gitlab) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error {
	filePath := paths.getPathComponents(iconName, iconfile.IconfileDescriptor).pathToIconfile
	commitErr := g.commit(modifiedBy, fmt.Sprintf("Adding iconfile: %s", filePath), []commitActionOnByteSlice{
		{
			Action:   commitActionCreate,
			FilePath: filePath,
			Content:  iconfile.Content,
		},
	})
	if commitErr != nil {
		return fmt.Errorf("failed to add iconfile to GitLab repo %s::%s: %w", iconName, iconfile.String(), commitErr)
	}
	g.logger.Info().Str("icon-name", iconName).Str("icon-file", iconfile.String()).Msg("Iconfile added to GitLab repository")
	return nil
}

func (g *Gitlab) DeleteIcon(iconDesc domain.IconDescriptor, modifiedBy authn.UserID) error {
	actionList := make([]commitActionOnByteSlice, len(iconDesc.Iconfiles))

	for index, ifDesc := range iconDesc.Iconfiles {
		actionList[index] = commitActionOnByteSlice{
			Action:   commitActionDelete,
			FilePath: paths.getPathComponents(iconDesc.Name, ifDesc).pathToIconfile,
		}
	}
	commitErr := g.commit(modifiedBy.String(), fmt.Sprintf("Deleting icon: %s", iconDesc.Name), actionList)
	if commitErr != nil {
		return fmt.Errorf("failed to delete iconfile from GitLab repo %s: %w", iconDesc.Name, commitErr)
	}
	g.logger.Info().Str("icon-name", iconDesc.Name).Msg("Iconfile deleted from GitLab repository")
	return nil
}

func (g *Gitlab) DeleteIconfile(iconName string, iconfileDesc domain.IconfileDescriptor, modifiedBy authn.UserID) error {
	filePath := paths.getPathComponents(iconName, iconfileDesc).pathToIconfile

	commitErr := g.commit(modifiedBy.String(), fmt.Sprintf("Deleting iconfile: %s", filePath), []commitActionOnByteSlice{
		{
			Action:   commitActionDelete,
			FilePath: filePath,
		},
	})
	if commitErr != nil {
		return fmt.Errorf("failed to delete iconfile from GitLab repo %s::%s: %w", iconName, iconfileDesc.String(), commitErr)
	}
	g.logger.Info().Str("icon-name", iconName).Str("icon-file", iconfileDesc.String()).Msg("Iconfile deleted from GitLab repository")
	return nil
}

func (g *Gitlab) GetIconfile(iconName string, iconfileDesc domain.IconfileDescriptor) ([]byte, error) {
	filePath := paths.getPathComponents(iconName, iconfileDesc).pathToIconfile
	statusCode, _, body, err := g.sendRequest(
		"GET",
		fmt.Sprintf(
			"/projects/%s/repository/files/%s?%s",
			url.PathEscape(g.project.String()),
			url.PathEscape(filePath),
			fmt.Sprintf("ref=%s", g.mainBranch),
		),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to get iconfigle from GitLab repo %s::%s: %w", iconName, iconfileDesc, err)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("failed to get iconfile from GitLab repo %s::%s: (%d) %s -- %w", iconName, iconfileDesc.String(), statusCode, body, err)
	}

	respFileItem := responseFileItem{}
	jsonErr := json.Unmarshal([]byte(body), &respFileItem)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal GitLab namespace list: %w", jsonErr)
	}

	if respFileItem.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding for iconfile from GitLab repo %s::%s: %s", iconName, iconfileDesc.String(), respFileItem.Encoding)
	}

	content, decodeErr := base64.StdEncoding.DecodeString(respFileItem.Content)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode iconfile content (%s) for %s::%s: %w", string(body), iconName, iconfileDesc.String(), decodeErr)
	}

	return content, nil
}

func (g *Gitlab) commit(authorName string, commitMessage string, actions []commitActionOnByteSlice) error {
	if os.Getenv(SimulateGitCommitFailureEnvvarName) == "true" {
		return fmt.Errorf("simulate git commit failure")
	}

	commitBody, createCommitBodyErr := g.createCommitBody(authorName, commitMessage, actions)
	if createCommitBodyErr != nil {
		return fmt.Errorf("failed to create commit request body: %w", createCommitBodyErr)
	}

	statusCode, _, body, err := g.sendRequest(
		"POST",
		fmt.Sprintf("/projects/%s/repository/commits?%s", url.PathEscape(g.project.String()), url.PathEscape(fmt.Sprintf("ref=%s", g.mainBranch))),
		commitBody,
	)
	if err != nil || statusCode != 201 {
		return fmt.Errorf("failed to commit to GitLab repo: (%d) %s -- %w", statusCode, body, err)
	}
	return nil
}

func (g *Gitlab) sendRequest(method string, apiCallPath string, body io.Reader) (int, http.Header, string, error) {
	urlString := fmt.Sprintf("https://gitlab.com/api/v4%s", apiCallPath)

	g.logger.Debug().Str("method", method).Str("url", urlString).Msg("send request")
	request, requestCreationError := http.NewRequest(
		method,
		urlString,
		body,
	)

	if requestCreationError != nil {
		return 0, nil, "", fmt.Errorf("failed to create request: %w", requestCreationError)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("PRIVATE-TOKEN", g.apikey)

	resp, requestExecutionError := g.client.Do(request)
	if requestExecutionError != nil {
		return 0, nil, "", fmt.Errorf("failed to execute request: %w", requestExecutionError)
	}
	defer resp.Body.Close()

	respBody, errBody := io.ReadAll(resp.Body)
	if errBody != nil {
		return resp.StatusCode, nil, "", fmt.Errorf("failed to read body: %w", errBody)
	}

	rateLimitRemainning, rateLimitParseErr := strconv.ParseInt(resp.Header.Get("RateLimit-Remaining"), 10, 0)
	if rateLimitParseErr != nil {
		return resp.StatusCode, nil, "", fmt.Errorf("failed to parse %s header", "RateLimit-Remaining")
	}
	if rateLimitRemainning < 5 {
		g.logger.Warn().Int64("rateLimitRemainning", rateLimitRemainning).Msg("Rate limit remaining to low")
	}

	return resp.StatusCode, resp.Header, string(respBody), nil
}

type namespaceInfo struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
}

func getNamespaceID(gitlabCli Gitlab) (int, error) {
	statusCode, _, body, err := gitlabCli.sendRequest("GET", "/namespaces?owned_only=true", nil)
	if err != nil || statusCode != 200 {
		return 0, fmt.Errorf("failed to retreive GitLab namespaces (%d) %s -- %w", statusCode, body, err)
	}

	namespaceInfoList := []namespaceInfo{}
	jsonErr := json.Unmarshal([]byte(body), &namespaceInfoList)
	if jsonErr != nil {
		return 0, fmt.Errorf("failed to unmarshal GitLab namespace list: %w", jsonErr)
	}

	for _, info := range namespaceInfoList {
		if info.Path == gitlabCli.project.namespacePath {
			return info.Id, nil
		}
	}

	return 0, fmt.Errorf("no namespace found with path %s", gitlabCli.project.namespacePath)
}
