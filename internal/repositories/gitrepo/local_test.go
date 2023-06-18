package gitrepo

import (
	"igo-repo/internal/logging"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const testRepoLocation = "test-git-repo"

var localGitRepoTestLogger = logging.CreateUnitLogger(logging.Get(), "repositories.localGitRepoTestSuite")

type localGitRepoTestSuite struct {
	suite.Suite
	t       *testing.T
	gitRepo Local
}

func TestLocalGitTestSuite(t *testing.T) {
	suite.Run(t, &localGitRepoTestSuite{t: t})
}

func (testSuite *localGitRepoTestSuite) removeRepoDir() {
	rmdirErr := os.RemoveAll(testRepoLocation)
	if rmdirErr != nil {
		panic(rmdirErr)
	}
}

func (testSuite *localGitRepoTestSuite) BeforeTest(suiteName string, testName string) {
	testSuite.removeRepoDir()
	gitRepo := NewLocalGitRepository(testRepoLocation)
	gitRepoCreationError := gitRepo.Create()
	if gitRepoCreationError != nil {
		panic(gitRepoCreationError)
	}
	testSuite.gitRepo = gitRepo
}

func (testSuite *localGitRepoTestSuite) TestLocationDoesntExist() {
	testSuite.removeRepoDir()
	gitRepo := Local{Location: testRepoLocation, Logger: localGitRepoTestLogger}
	testSuite.Equal(false, gitRepo.locationHasRepo())
}

func (testSuite *localGitRepoTestSuite) TestLocationDoesntHaveRepo() {
	testSuite.removeRepoDir()
	gitRepo := NewLocalGitRepository(testRepoLocation)
	testSuite.Equal(false, gitRepo.locationHasRepo())
}

func (testSuite *localGitRepoTestSuite) TestLocationHasRepo() {
	testSuite.removeRepoDir()
	gitRepo := NewLocalGitRepository(testRepoLocation)
	err := gitRepo.Delete()
	if err != nil {
		panic(err)
	}
	err = gitRepo.Create()
	if err != nil {
		panic(err)
	}
	testSuite.Equal(true, gitRepo.locationHasRepo())
}

func (testSuite *localGitRepoTestSuite) TestParseCommitMetadata() {
	// git show --quiet --format=fuller --date=format:'%Y-%m-%dT%H:%M:%S%z' 95fb34a325e697bafffb785ac65ecca986ca06a6

	testInput := `Author:     Kovács, Péter <peter.dunay.kovacs@gmail.com>
AuthorDate: 2022-10-09T13:42:12+0200
Commit:     Péter Kovács <peter.dunay.kovacs@gmail.com>
CommitDate: 2022-10-31T15:30:17+0100
	
    [DEV] distributed git access`

	authorDate, authorDateErr := time.Parse(time.RFC3339, "2022-10-09T13:42:12+02:00")
	testSuite.Nil(authorDateErr)
	commitDate, commitDateErr := time.Parse(time.RFC3339, "2022-10-31T15:30:17+01:00")
	testSuite.Nil(commitDateErr)

	expectedOutput := CommitMetadata{
		Author:     "Kovács, Péter <peter.dunay.kovacs@gmail.com>",
		AuthorDate: authorDate,
		Commit:     "Péter Kovács <peter.dunay.kovacs@gmail.com>",
		CommitDate: commitDate,
		Message:    "[DEV] distributed git access",
	}

	commitMetadata, parseErr := parseLocalCommitMetadata(testInput)
	testSuite.Nil(parseErr)
	testSuite.Equal(expectedOutput, commitMetadata)
}
