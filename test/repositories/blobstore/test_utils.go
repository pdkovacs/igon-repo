package blobstore

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/blobstore/git"
	git_tests "iconrepo/test/repositories/blobstore/git"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type TestBlobstoreClientFactory func(conf *config.Options) (TestBlobstoreClient, error)

type blobstoreManagement interface {
	ResetRepository() error
	DeleteRepository() error
	CheckStatus() (bool, error)
	GetStateID() (string, error)
	GetIconfiles() ([]string, error)
	GetIconfile(iconName string, iconfileDesc domain.IconfileDescriptor) ([]byte, error)
	GetVersionFor(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error)
	GetVersionMetadata(commitId string) (git.CommitMetadata, error)
}

type TestBlobstoreClient interface {
	repositories.BlobstoreRepository
	blobstoreManagement
}

type TestBlobstoreController struct {
	repoFactory TestBlobstoreClientFactory
	repo        TestBlobstoreClient
}

func (ctl *TestBlobstoreController) String() string {
	return ctl.repo.String()
}

func (ctl *TestBlobstoreController) ResetRepository(conf *config.Options) error {
	if ctl.repo != nil {
		ctl.repo.DeleteRepository()
	}
	var err error
	ctl.repo, err = ctl.repoFactory(conf)
	if err != nil {
		return err
	}
	return ctl.repo.CreateRepository()
}

func (ctl *TestBlobstoreController) DeleteRepository() error {
	return ctl.repo.DeleteRepository()
}

func (ctl *TestBlobstoreController) GetStateID() (string, error) {
	return ctl.repo.GetStateID()
}

func (ctl *TestBlobstoreController) CheckStatus() (bool, error) {
	return ctl.repo.CheckStatus()
}

func (ctl *TestBlobstoreController) GetVersionFor(iconName string, iconfileDesc domain.IconfileDescriptor) (string, error) {
	return ctl.repo.GetVersionFor(iconName, iconfileDesc)
}

func (ctl *TestBlobstoreController) GetVersionMetadata(commitId string) (git.CommitMetadata, error) {
	return ctl.repo.GetVersionMetadata(commitId)
}

func (ctl *TestBlobstoreController) GetIconfiles() ([]string, error) {
	return ctl.repo.GetIconfiles()
}

func (ctl *TestBlobstoreController) GetIconfile(iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	return ctl.repo.GetIconfile(iconName, iconfile)
}

func (ctl *TestBlobstoreController) AddIconfile(iconName string, iconfile domain.Iconfile, modifiedBy string) error {
	return ctl.repo.AddIconfile(iconName, iconfile, modifiedBy)
}

type BlobstoreTestSuite struct {
	suite.Suite
	RepoController TestBlobstoreController
	TestSequenceId string
	TestCaseId     int
}

func (s *BlobstoreTestSuite) BeforeTest(suiteName, testName string) {
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	s.TestCaseId++
	git_tests.SetupGitlabTestCaseConfig(&conf, s.TestSequenceId, strconv.Itoa(s.TestCaseId))
	createRepoErr := s.RepoController.ResetRepository(&conf)
	if createRepoErr != nil {
		s.FailNow("", "%v", createRepoErr)
	}
}

func (s *BlobstoreTestSuite) AfterTest(suiteName, testName string) {
	os.Unsetenv(git.SimulateGitCommitFailureEnvvarName)
	err := s.RepoController.DeleteRepository()
	if err != nil {
		s.FailNow("", "%v", err)
	}
}

func (s *BlobstoreTestSuite) GetStateID() (string, error) {
	return s.RepoController.GetStateID()
}

func (s *BlobstoreTestSuite) AssertBlobstoreCleanStatus() {
	status, err := s.RepoController.CheckStatus()
	s.NoError(err)
	s.True(status)
}

func (s *BlobstoreTestSuite) AssertFileInBlobstore(iconName string, iconfile domain.Iconfile, timeBeforeCommit time.Time) {
	commitID, getCommitIDErr := s.RepoController.GetVersionFor(iconName, iconfile.IconfileDescriptor)
	s.NoError(getCommitIDErr)
	s.Greater(len(commitID), 0)
	meta, getMetaErr := s.RepoController.GetVersionMetadata(commitID)
	s.NoError(getMetaErr)
	s.Greater(meta.CommitDate, timeBeforeCommit.Add(-time.Duration(1_000)*time.Millisecond))
}

func (s *BlobstoreTestSuite) AssertFileNotInBlobstore(iconName string, iconfile domain.Iconfile) {
	commitId, err := s.RepoController.GetVersionFor(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal("", commitId)
}

func NewLocalGitTestRepo(conf *config.Options) (*git.Local, error) {
	conf.GitlabNamespacePath = "" // to guide the test app on which git provider to used
	repo := git.NewLocalGitRepository(conf.LocalGitRepo, logging.CreateUnitLogger(logging.Get(), "test local git repository"))
	return &repo, nil
}

var DefaultBlobstoreController = TestBlobstoreController{
	repoFactory: func(conf *config.Options) (TestBlobstoreClient, error) {
		return NewLocalGitTestRepo(conf)
	},
}

func BlobstoreProvidersToTest() []TestBlobstoreController {

	fmt.Printf(">>>>>>>>>>>> LOCAL_GIT_ONLY: %v\n", os.Getenv("LOCAL_GIT_ONLY"))
	if len(os.Getenv("LOCAL_GIT_ONLY")) > 0 {
		return []TestBlobstoreController{DefaultBlobstoreController}
	}

	return []TestBlobstoreController{
		DefaultBlobstoreController,
		{
			repoFactory: func(conf *config.Options) (TestBlobstoreClient, error) {
				repo, createClientErr := git_tests.NewGitlabTestRepoClient(conf)
				if createClientErr != nil {
					return nil, createClientErr
				}
				return repo, nil
			},
		},
	}
}
