package blobstore

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"iconrepo/internal/app/domain"
	"iconrepo/internal/config"
	"iconrepo/internal/repositories"
	"iconrepo/internal/repositories/blobstore/git"
	git_tests "iconrepo/test/repositories/blobstore/git"
	"iconrepo/test/test_commons"

	"github.com/stretchr/testify/suite"
)

type TestBlobstoreClientFactory func(conf *config.Options) (TestBlobstoreClient, error)

type blobstoreManagement interface {
	ResetRepository(ctx context.Context) error
	DeleteRepository(ctx context.Context) error
	CheckStatus() (bool, error)
	GetStateID(ctx context.Context) (string, error)
	GetIconfiles(ctx context.Context) ([]string, error)
	GetVersionFor(ctx context.Context, iconName string, iconfileDesc domain.IconfileDescriptor) (string, error)
	GetVersionMetadata(ctx context.Context, commitId string) (git.CommitMetadata, error)
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

func (ctl *TestBlobstoreController) ResetRepository(ctx context.Context, conf *config.Options) error {
	if ctl.repo != nil {
		ctl.repo.DeleteRepository(ctx)
	}
	var err error
	ctl.repo, err = ctl.repoFactory(conf)
	if err != nil {
		return err
	}
	return ctl.repo.CreateRepository(ctx)
}

func (ctl *TestBlobstoreController) DeleteRepository(ctx context.Context) error {
	return ctl.repo.DeleteRepository(ctx)
}

func (ctl *TestBlobstoreController) GetStateID(ctx context.Context) (string, error) {
	return ctl.repo.GetStateID(ctx)
}

func (ctl *TestBlobstoreController) CheckStatus() (bool, error) {
	return ctl.repo.CheckStatus()
}

func (ctl *TestBlobstoreController) GetVersionFor(ctx context.Context, iconName string, iconfileDesc domain.IconfileDescriptor) (string, error) {
	return ctl.repo.GetVersionFor(ctx, iconName, iconfileDesc)
}

func (ctl *TestBlobstoreController) GetVersionMetadata(ctx context.Context, commitId string) (git.CommitMetadata, error) {
	return ctl.repo.GetVersionMetadata(ctx, commitId)
}

func (ctl *TestBlobstoreController) GetIconfiles(ctx context.Context) ([]string, error) {
	return ctl.repo.GetIconfiles(ctx)
}

func (ctl *TestBlobstoreController) GetIconfile(ctx context.Context, iconName string, iconfile domain.IconfileDescriptor) ([]byte, error) {
	return ctl.repo.GetIconfile(ctx, iconName, iconfile)
}

func (ctl *TestBlobstoreController) AddIconfile(ctx context.Context, iconName string, iconfile domain.Iconfile, modifiedBy string) error {
	return ctl.repo.AddIconfile(ctx, iconName, iconfile, modifiedBy)
}

type BlobstoreTestSuite struct {
	suite.Suite
	RepoController TestBlobstoreController
	TestSequenceId string
	TestCaseId     int
	Ctx            context.Context
}

func (s *BlobstoreTestSuite) BeforeTest(suiteName, testName string) {
	conf := test_commons.CloneConfig(test_commons.GetTestConfig())
	s.TestCaseId++
	git_tests.SetupGitlabTestCaseConfig(&conf, s.TestSequenceId, strconv.Itoa(s.TestCaseId))
	createRepoErr := s.RepoController.ResetRepository(s.Ctx, &conf)
	if createRepoErr != nil {
		s.FailNow("", "%v", createRepoErr)
	}
}

func (s *BlobstoreTestSuite) AfterTest(suiteName, testName string) {
	os.Unsetenv(git.SimulateGitCommitFailureEnvvarName)
	err := s.RepoController.DeleteRepository(s.Ctx)
	if err != nil {
		s.FailNow("", "%v", err)
	}
}

func (s *BlobstoreTestSuite) GetStateID() (string, error) {
	return s.RepoController.GetStateID(s.Ctx)
}

func (s *BlobstoreTestSuite) AssertBlobstoreCleanStatus() {
	status, err := s.RepoController.CheckStatus()
	s.NoError(err)
	s.True(status)
}

func (s *BlobstoreTestSuite) AssertFileInBlobstore(iconName string, iconfile domain.Iconfile, timeBeforeCommit time.Time) {
	commitID, getCommitIDErr := s.RepoController.GetVersionFor(s.Ctx, iconName, iconfile.IconfileDescriptor)
	s.NoError(getCommitIDErr)
	s.Greater(len(commitID), 0)
	meta, getMetaErr := s.RepoController.GetVersionMetadata(s.Ctx, commitID)
	s.NoError(getMetaErr)
	s.Greater(meta.CommitDate, timeBeforeCommit.Add(-time.Duration(1_000)*time.Millisecond))
}

func (s *BlobstoreTestSuite) AssertFileNotInBlobstore(iconName string, iconfile domain.Iconfile) {
	commitId, err := s.RepoController.GetVersionFor(s.Ctx, iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal("", commitId)
}

func NewLocalGitTestRepo(conf *config.Options) (*git.Local, error) {
	conf.GitlabNamespacePath = "" // to guide the test app on which git provider to used
	repo := git.NewLocalGitRepository(conf.LocalGitRepo)
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
