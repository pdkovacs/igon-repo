package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
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

type readConfigurationTestSuite struct {
	suite.Suite
}

func TestReadConfiguration(t *testing.T) {
	suite.Run(t, &readConfigurationTestSuite{})
}

func createTempFileForConfig() *os.File {
	logger := logrus.WithField("prefix", "createTempFileForConfig")
	file, tempFileError := ioutil.TempFile(os.Getenv("HOME"), "TestReadConfiguration")
	if tempFileError != nil {
		logger.Fatal(tempFileError)
	}
	return file
}

func storeConfigInTempFile(config interface{}, file *os.File) {
	logger := logrus.WithField("prefix", "storeConfigInTempFile")
	wireForm, marshallingError := json.Marshal(config)
	if marshallingError != nil {
		logger.Fatal(marshallingError)
	}
	countWritten, writeError := file.Write(wireForm)
	if writeError != nil {
		logger.Fatal(writeError)
	}
	if countWritten != len(wireForm) {
		logger.Fatalf("Couldn't write all config: %d bytes instead of %d", countWritten, len(wireForm))
	}
}

func (s *readConfigurationTestSuite) AfterTest(suiteName, testName string) {
	logrus.Infof("AfterTest called")
	clearEnvVarsSet()
}

func (s *readConfigurationTestSuite) TestYieldDefaultsWithoutConfigFile() {
	clArgs := []string{}
	opts := ReadConfiguration("some non-existent file", clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal("localhost", opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestConfigFileSettingsOverridesDefaults() {
	connHostInFile := "tohuvabohu"
	optsInFile := make(map[string]interface{})
	optsInFile["dbHost"] = connHostInFile

	configFile := createTempFileForConfig()
	// defer os.Remove(configFile.Name())
	storeConfigInTempFile(optsInFile, configFile)

	clArgs := []string{}
	opts := ReadConfiguration(configFile.Name(), clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(connHostInFile, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestEnvVarSettingOverridesDefaults() {
	connHostInEnvVar := "nokedli"

	setEnvVar("DB_HOST", connHostInEnvVar)

	clArgs := []string{}
	opts := ReadConfiguration("name-of-some-nonexistent-file", clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(connHostInEnvVar, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestCliArgsOverridesEnvVarSettings() {
	connHostInEnvVar := "nokedli"
	connHostInArg := "tohuvabohu"
	optsInFile := Options{}
	optsInFile.DBHost = connHostInArg

	setEnvVar("DB_HOST", connHostInEnvVar)

	clArgs := []string{"--db-host", connHostInArg}
	opts := ReadConfiguration("name-of-some-nonexistent-file", clArgs)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(connHostInArg, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}
