package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

var envVarsSet []string = []string{}

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
	logger := log.WithField("prefix", "test-closeRemoveFile")
	fileCloseError := file.Close()
	if fileCloseError != nil {
		logger.Errorf("Error while closing configuration file %v: %v", file.Name(), fileCloseError)
	}
	fileRemoveError := os.Remove(file.Name())
	if fileRemoveError != nil {
		logger.Errorf("Error while closing configuration file %v: %v", file.Name(), fileRemoveError)
	}
}

type readConfigurationTestSuite struct {
	suite.Suite
}

func TestReadConfiguration(t *testing.T) {
	suite.Run(t, &readConfigurationTestSuite{})
}

func createTempFileForConfig() *os.File {
	logger := log.WithField("prefix", "createTempFileForConfig")
	file, tempFileError := ioutil.TempFile(os.Getenv("HOME"), "TestReadConfiguration")
	if tempFileError != nil {
		logger.Fatal(tempFileError)
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

func storeConfigInTempFile(key string, value string) (configFile *os.File) {
	var err *error
	configFile = createTempFileForConfig()
	defer func() {
		if *err != nil {
			log.Errorf("Error while storing config file: %v", *err)
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
	opts := GetDefaultConfiguration()
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal("localhost", opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
	s.Equal(DefaultIconDataLocationGit, opts.IconDataLocationGit)
}

func (s *readConfigurationTestSuite) TestFailOnMissingConfigFile() {
	clArgs := []string{}
	_, err := ReadConfiguration("some non-existent file", clArgs)
	s.True(errors.Is(err, fs.ErrNotExist))
}

// TODO: "mergo" doesn't overwrite non-empty values
// The simplest fix seems to be handling defaults outside github.com/jessevdk/go-flags as a third step:
// 1. go-flags is configured to use empty defaults
// 2. config.json is overwritten with the go-flags output
// 3. empty-defaults are overwritten with non-empty defaults (where to store the non-empties???)
func (s *readConfigurationTestSuite) /* Test */ ConfigFileSettingsOverridesDefaults() {
	dbHostInFile := "tohuvabohu"

	configFile := storeConfigInTempFile("dbHost", dbHostInFile)
	defer closeRemoveFile(configFile)

	clArgs := []string{}
	opts, err := ReadConfiguration(configFile.Name(), clArgs)
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
	opts := ParseCommandLineArgs(clArgs)
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
	opts, err := ReadConfiguration(configFile.Name(), clArgs)
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
	opts, err := ReadConfiguration(configFile.Name(), clArgs)
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
	opts := ParseCommandLineArgs(clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInArg, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}
