package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"igo-repo/internal/config"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

var envVarsSet []string = []string{}

var unitLogID = "test-closeRemoveFile"

func clearEnvVarsSet() {
	for _, envVar := range envVarsSet {
		os.Unsetenv(envVar)
	}
	envVarsSet = []string{}
}

func setEnvVar(envVarName string, value string) {
	envVarsSet = append(envVarsSet, envVarName)
	os.Setenv(envVarName, value)
}

func closeRemoveFile(file *os.File) {
	fileCloseError := file.Close()
	if fileCloseError != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error while closing configuration file %v: %v\n", unitLogID, file.Name(), fileCloseError)
	}
	fileRemoveError := os.Remove(file.Name())
	if fileRemoveError != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error while closing configuration file %v: %v\n", unitLogID, file.Name(), fileRemoveError)
	}
}

type readConfigurationTestSuite struct {
	suite.Suite
}

func TestReadConfiguration(t *testing.T) {
	suite.Run(t, &readConfigurationTestSuite{})
}

func createTempFileForConfig() *os.File {
	file, tempFileError := ioutil.TempFile(os.Getenv("HOME"), "TestReadConfiguration")
	if tempFileError != nil {
		panic(tempFileError)
	}
	return file
}

func storeJSONConfig(config map[string]interface{}, file *os.File) error {
	wireForm, marshallingError := json.Marshal(config)
	if marshallingError != nil {
		return fmt.Errorf("Failed to marshal the configuration into JSON: %w", marshallingError)
	}
	countWritten, writeError := file.Write(wireForm)
	if writeError != nil {
		return fmt.Errorf("Failed to write the configuration JSON to file %s: %w", file.Name(), writeError)
	}
	if countWritten != len(wireForm) {
		return fmt.Errorf("Failed to write the configuration JSON to file %s: wrote %d bytes instead of %d", file.Name(), countWritten, len(wireForm))
	}
	return nil
}

func storeConfigInTempFile(key string, value interface{}) (configFile *os.File) {
	var err *error
	configFile = createTempFileForConfig()
	defer func() {
		if *err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Error while storing config file: %v\n", unitLogID, *err)
			closeRemoveFile(configFile)
			panic(err)
		}
	}()

	optsInFile := make(map[string]interface{})
	optsInFile[key] = value
	storeError := storeJSONConfig(optsInFile, configFile)
	err = &storeError
	return
}

func (s *readConfigurationTestSuite) AfterTest(suiteName, testName string) {
	clearEnvVarsSet()
}

func (s *readConfigurationTestSuite) TestGetDefaultConfiguration() {
	opts := config.GetDefaultConfiguration()
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal("localhost", opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
	s.Equal(config.DefaultIconDataLocationGit, opts.IconDataLocationGit)
}

func (s *readConfigurationTestSuite) TestFailOnMissingConfigFile() {
	clArgs := []string{}
	_, err := config.ReadConfiguration("some non-existent file", clArgs)
	s.True(errors.Is(err, fs.ErrNotExist))
}

// TODO_MAYBE: "mergo" doesn't overwrite non-empty values
// The simplest fix seems to be handling defaults outside github.com/jessevdk/go-flags as a third step:
// 1. go-flags is configured to use empty defaults
// 2. config.json is overwritten with the go-flags output
// 3. empty-defaults are overwritten with non-empty defaults (where to store the non-empties???)
// Currently, only simple configuration setting values have default, which are easy to override from the CL or via env var
func (s *readConfigurationTestSuite) TestConfigFileSettingsOverridesDefaults() {
	dbHostInFile := "tohuvabohu"

	configFile := storeConfigInTempFile("dbHost", dbHostInFile)
	defer closeRemoveFile(configFile)

	clArgs := []string{}
	opts, err := config.ReadConfiguration(configFile.Name(), clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInFile, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestEnvVarSettingOverridesDefaults() {
	dbHostInEnvVar := "nokedli"

	setEnvVar("DB_HOST", dbHostInEnvVar)

	clArgs := []string{}
	opts := config.ParseCommandLineArgs(clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInEnvVar, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestEnvVarSettingOverridesConfigFile() {
	dbHostInFile := "tohuvabohu"
	dbHostInEnvVar := "nokedli"

	configFile := storeConfigInTempFile("dbHost", dbHostInFile)
	defer closeRemoveFile(configFile)

	setEnvVar("DB_HOST", dbHostInEnvVar)

	clArgs := []string{}
	opts, err := config.ReadConfiguration(configFile.Name(), clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInEnvVar, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestCliArgsOverrideConfigFile() {
	dbHostInFile := "tohuvabohu"
	dbHostInArg := "nokedli"

	configFile := storeConfigInTempFile("dbHost", dbHostInFile)
	defer closeRemoveFile(configFile)

	clArgs := []string{"--db-host", dbHostInArg}
	opts, err := config.ReadConfiguration(configFile.Name(), clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInArg, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestCliArgsOverridesEnvVarSettings() {
	connHostInEnvVar := "nokedli"
	dbHostInArg := "tohuvabohu"

	setEnvVar("DB_HOST", connHostInEnvVar)

	clArgs := []string{"--db-host", dbHostInArg}
	opts := config.ParseCommandLineArgs(clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInArg, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestPasswordCredentialsFromConfigFile() {
	expected := []config.PasswordCredentials{{
		Username: "zazi",
		Password: "metro",
	}}

	configFile := storeConfigInTempFile("passwordCredentials", expected)
	defer closeRemoveFile(configFile)

	clArgs := []string{}
	opts, err := config.ReadConfiguration(configFile.Name(), clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(expected, opts.PasswordCredentials)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestUsersByRolesFromConfigFile() {
	expected := config.UsersByRoles{"zazi": []string{"metro", "paris"}}

	configFile := storeConfigInTempFile("usersByRoles", expected)
	defer closeRemoveFile(configFile)

	clArgs := []string{}
	opts, err := config.ReadConfiguration(configFile.Name(), clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(expected, opts.UsersByRoles)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}
