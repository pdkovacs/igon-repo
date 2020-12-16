package server

import (
	"encoding/json"
	"fmt"
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

func closeRemoveFile(file *os.File) {
	fileCloseError := file.Close()
	if fileCloseError != nil {
		logrus.Errorf("Error while closing configuration file %v: %v", file.Name(), fileCloseError)
	}
	fileRemoveError := os.Remove(file.Name())
	if fileRemoveError != nil {
		logrus.Errorf("Error while closing configuration file %v: %v", file.Name(), fileRemoveError)
	}
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

func storeJSONConfig(config map[string]interface{}, file *os.File) error {
	wireForm, marshallingError := json.Marshal(config)
	if marshallingError != nil {
		return marshallingError
	}
	countWritten, writeError := file.Write(wireForm)
	if writeError != nil {
		return writeError
	}
	if countWritten != len(wireForm) {
		return fmt.Errorf("Couldn't write all config: %d bytes instead of %d", countWritten, len(wireForm))
	}
	return nil
}

func storeConfigInTempFile(key string, value string) (configFile *os.File, err error) {
	configFile = createTempFileForConfig()
	defer func() {
		logrus.Infof("aaa error: %v", err)
		if err != nil {
			closeRemoveFile(configFile)
		}
	}()

	optsInFile := make(map[string]interface{})
	optsInFile[key] = value
	err = storeJSONConfig(optsInFile, configFile)
	logrus.Infof("ccc error: %v", err)
	return
}

func (s *readConfigurationTestSuite) AfterTest(suiteName, testName string) {
	logrus.Infof("AfterTest called")
	clearEnvVarsSet()
}

func (s *readConfigurationTestSuite) TestYieldDefaultsWithoutConfigFile() {
	clArgs := []string{}
	opts, err := ReadConfiguration("some non-existent file", clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal("localhost", opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestConfigFileSettingsOverridesDefaults() {
	dbHostInFile := "tohuvabohu"

	configFile, err := storeConfigInTempFile("dbHost", dbHostInFile)
	if err != nil {
		panic(err)
	}
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
	opts, err := ReadConfiguration("name-of-some-nonexistent-file", clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInEnvVar, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}

func (s *readConfigurationTestSuite) TestEnvVarSettingOverridesConfigFile() {
	dbHostInFile := "tohuvabohu"
	dbHostInEnvVar := "nokedli"

	configFile, err := storeConfigInTempFile("dbHost", dbHostInFile)
	if err != nil {
		panic(err)
	}
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

	configFile, err := storeConfigInTempFile("dbHost", dbHostInFile)
	if err != nil {
		panic(err)
	}
	defer closeRemoveFile(configFile)

	clArgs := []string{"--db-host", dbHostInArg}
	opts, err := ReadConfiguration("name-of-some-nonexistent-file", clArgs)
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
	opts, err := ReadConfiguration("name-of-some-nonexistent-file", clArgs)
	s.NoError(err)
	s.Equal("localhost", opts.ServerHostname)
	s.Equal(8080, opts.ServerPort)
	s.Equal(dbHostInArg, opts.DBHost)
	s.Equal(5432, opts.DBPort)
	s.Equal(false, opts.EnableBackdoors)
}
