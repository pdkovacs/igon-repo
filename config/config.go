package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const BasicAuthentication = "basic"
const OIDCAuthentication = "oidc"

// PasswordCredentials holds password-credentials
type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UsersByRoles maps roles to lists of user holding the role
type UsersByRoles map[string][]string

// Options holds the available command-line options
type Options struct {
	ServerHostname              string                `json:"serverHostname" env:"SERVER_HOSTNAME" long:"server-hostname" short:"h" default:"localhost" description:"Server hostname"`
	ServerPort                  int                   `json:"serverPort" env:"SERVER_PORT" long:"server-port" short:"p" default:"8080" description:"Server port"`
	ServerURLContext            string                `json:"serverUrlContext" env:"SERVER_URL_CONTEXT" long:"server-url-context" short:"c" default:"" description:"Server url context"`
	AppDescription              string                `json:"appDescription" env:"APP_DESCRIPTION" long:"app-description" short:"" default:"" description:"Application description"`
	IconDataLocationGit         string                `json:"iconDataLocationGit" env:"ICON_DATA_LOCATION_GIT" long:"icon-data-location-git" short:"g" default:"" description:"Icon data location git"`
	IconDataCreateNew           string                `json:"iconDataCreateNew" env:"ICON_DATA_CREATE_NEW" long:"icon-data-create-new" short:"n" default:"never" description:"Icon data create new"`
	AuthenticationType          string                `json:"authenticationType" env:"AUTHENTICATION_TYPE" long:"authentication-type" short:"a" default:"oidc" description:"Authentication type"`
	PasswordCredentials         []PasswordCredentials `json:"passwordCredentials" env:"PASSWORD_CREDENTIALS" long:"password-credentials"`
	OIDCClientID                string                `json:"oidcClientId" env:"OIDC_CLIENT_ID" long:"oidc-client-id" short:"" default:"" description:"OIDC client id"`
	OIDCClientSecret            string                `json:"oidcClientSecret" env:"OIDC_CLIENT_SECRET" long:"oidc-client-secret" short:"" default:"" description:"OIDC client secret"`
	OIDCAccessTokenURL          string                `json:"oidcAccessTokenUrl" env:"OIDC_ACCESS_TOKEN_URL" long:"oidc-access-token-url" short:"" default:"" description:"OIDC access token url"`
	OIDCUserAuthorizationURL    string                `json:"oidcUserAuthorizationUrl" env:"OIDC_USER_AUTHORIZATION_URL" long:"oidc-user-authorization-url" short:"" default:"" description:"OIDC user authorization url"`
	OIDCClientRedirectBackURL   string                `json:"oidcClientRedirectBackUrl" env:"OIDC_CLIENT_REDIRECT_BACK_URL" long:"oidc-client-redirect-back-url" short:"" default:"" description:"OIDC client redirect back url"`
	OIDCTokenIssuer             string                `json:"oidcTokenIssuer" env:"OIDC_TOKEN_ISSUER" long:"oidc-token-issuer" short:"" default:"" description:"OIDC token issuer"`
	OIDCIpJwtPublicKeyURL       string                `json:"oidcIpJwtPublicKeyUrl" env:"OIDC_IP_JWT_PUBLIC_KEY_URL" long:"oidc-ip-jwt-public-key-url" short:"" default:"" description:"OIDC ip jwt public key url"`
	OIDCIpJwtPublicKeyPemBase64 string                `json:"oidcIpJwtPublicKeyPemBase64" env:"OIDC_IP_JWT_PUBLIC_KEY_PEM_BASE64" long:"oidc-ip-jwt-public-key-pem-base64" short:"" default:"" description:"OIDC ip jwt public key pem base64"`
	OIDCIpLogoutURL             string                `json:"oidcIpLogoutUrl" env:"OIDC_IP_LOGOUT_URL" long:"oidc-ip-logout-url" short:"" default:"" description:"OIDC ip logout url"`
	UsersByRoles                UsersByRoles          `json:"usersByRoles" env:"USERS_BY_ROLES" long:"users-by-roles" short:"" default:"" description:"Users by roles"`
	DBHost                      string                `json:"dbHost" env:"DB_HOST" long:"db-host" short:"" default:"localhost" description:"DB host"`
	DBPort                      int                   `json:"dbPort" env:"DB_PORT" long:"db-port" short:"" default:"5432" description:"DB port"`
	DBUser                      string                `json:"dbUser" env:"DB_USER" long:"db-user" short:"" default:"iconrepo" description:"DB user"`
	DBPassword                  string                `json:"dbPassword" env:"DB_PASSWORD" long:"db-password" short:"" default:"iconrepo" description:"DB password"`
	DBName                      string                `json:"dbName" env:"DB_NAME" long:"db-name" short:"" default:"iconrepo" description:"Name of the database"`
	DBSchemaName                string                `json:"dbSchemaName" env:"DB_SCHEMA_NAME" long:"db-schema-name" short:"" default:"icon_repo" description:"Name of the database schemma"`
	EnableBackdoors             bool                  `json:"enableBackdoors" env:"ENABLE_BACKDOORS" long:"enable-backdoors" short:"" description:"Enable backdoors"`
	PackageRootDir              string                `json:"packageRootDir" env:"PACKAGE_ROOT_DIR" long:"package-root-dir" short:"" default:"" description:"Package root dir"`
	LogLevel                    string                `json:"logLevel" env:"IGOREPO_LOG_LEVEL" long:"log-level" short:"l" default:"info"`
}

var DefaultIconRepoHome = filepath.Join(os.Getenv("HOME"), ".ui-toolbox/icon-repo")
var DefaultIconDataLocationGit = filepath.Join(DefaultIconRepoHome, "git-repo")
var DefaultConfigFilePath = filepath.Join(DefaultIconRepoHome, "config.json")

// GetConfigFilePath gets the path of the configuration file
func GetConfigFilePath() string {
	var result string
	if result = os.Getenv("ICON_REPO_CONFIG_FILE"); result != "" {
	} else {
		result = DefaultConfigFilePath
	}
	log.Infof("Configuration file: %s", result)
	return result
}

// ReadConfiguration reads the configuration file and merges it with the command line arguments
func ReadConfiguration(filePath string, clArgs []string) (Options, error) {
	mapInFile, optsInFile, err := readConfigurationFromFile(filePath)
	if err != nil {
		return Options{}, err
	}
	return parseFlagsMergeSettings(clArgs, mapInFile, optsInFile), nil
}

func readConfigurationFromFile(filePath string) (map[string]interface{}, Options, error) {
	var mapInFile = make(map[string]interface{})
	optsInFile := Options{}

	_, fileStatError := os.Stat(filePath)
	if fileStatError != nil {
		return mapInFile, optsInFile, fmt.Errorf("failed to locate configuration file %v: %w", filePath, fileStatError)
	}

	fileContent, fileReadError := os.ReadFile(filePath)
	if fileReadError != nil {
		return mapInFile, optsInFile, fmt.Errorf("failed to read configuration file %v: %w", filePath, fileReadError)
	}

	unmarshalError := json.Unmarshal(fileContent, &mapInFile)
	if unmarshalError != nil {
		return mapInFile, optsInFile, fmt.Errorf("failed to parse configuration file %v: %w", filePath, unmarshalError)
	}

	unmarshalError1 := json.Unmarshal(fileContent, &optsInFile)
	if unmarshalError1 != nil {
		return mapInFile, optsInFile, fmt.Errorf("failed to parse configuration file %v: %w", filePath, unmarshalError1)
	}

	return mapInFile, optsInFile, nil
}

func GetDefaultConfiguration() Options {
	options := parseFlagsMergeSettings([]string{}, nil, Options{})
	return options
}

func ParseCommandLineArgs(clArgs []string) Options {
	options := parseFlagsMergeSettings(clArgs, nil, Options{})
	return options
}

func findCliArg(name string, clArgs []string) string {
	for index, arg := range clArgs {
		if arg == name {
			return clArgs[index+1]
		}
	}
	return ""
}

func parseFlagsMergeSettings(clArgs []string, mapFromConfFile map[string]interface{}, optsInConfFile Options) Options {
	opts := Options{}
	psResult := reflect.ValueOf(&opts)
	t := reflect.TypeOf(opts)
	for fieldIndex := 0; fieldIndex < t.NumField(); fieldIndex++ {
		field := t.FieldByIndex([]int{fieldIndex})
		fWritable := psResult.Elem().FieldByName(field.Name)
		if !fWritable.IsValid() || !fWritable.CanSet() {
			panic(fmt.Sprintf("Field %v is not valid (%v) or not writeable (%v).", field.Name, fWritable.IsValid(), fWritable.CanAddr()))
		}
		longopt := fmt.Sprintf("--%s", field.Tag.Get("long"))
		if cliArg := findCliArg(longopt, clArgs); cliArg != "" {
			err := setValueFromString(cliArg, fWritable)
			if err != nil {
				panic(fmt.Errorf("failed to set command-line argument value %v=%v: %w", longopt, cliArg, err))
			}
			continue
		}
		envName := field.Tag.Get("env")
		if envVal := os.Getenv(envName); envVal != "" {
			err := setValueFromString(envVal, fWritable)
			if err != nil {
				panic(fmt.Errorf("failed to set environment variable value %v=%v: %w", envName, envVal, err))
			}
			continue
		}
		if mapFromConfFile != nil {
			propName := field.Tag.Get("json")
			if value, ok := mapFromConfFile[propName]; ok {
				err := setValueFromInterface(value, fWritable)
				if err != nil {
					panic(fmt.Errorf("failed to set config file value %v=%v: %w", envName, value, err))
				}
				continue
			}
		}
		if dfltValue := field.Tag.Get("default"); dfltValue != "" {
			err := setValueFromString(dfltValue, fWritable)
			if err != nil {
				panic(fmt.Errorf("failed to set default value %v=%v: %w", envName, dfltValue, err))
			}
			continue
		}
	}

	opts.PasswordCredentials = optsInConfFile.PasswordCredentials
	opts.UsersByRoles = optsInConfFile.UsersByRoles

	if opts.IconDataLocationGit == "" {
		opts.IconDataLocationGit = DefaultIconDataLocationGit
	}

	return opts
}

func setValueFromString(value string, target reflect.Value) error {
	switch target.Kind() {
	case reflect.Int:
		{
			val, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			x := int64(val)
			if target.OverflowInt(x) {
				return fmt.Errorf("overflow error")
			}
			target.SetInt(x)
		}
	case reflect.String:
		{
			target.SetString(value)
		}
	default:
		return fmt.Errorf("unexpected property type: %v", target.Kind())
	}
	return nil
}

func setValueFromInterface(value interface{}, target reflect.Value) error {
	switch target.Kind() {
	case reflect.Int:
		{
			if val, ok := value.(int); ok {
				x := int64(val)
				if target.OverflowInt(x) {
					return fmt.Errorf("overflow error")
				}
				target.SetInt(x)
				return nil
			}
			if val, ok := value.(float64); ok {
				x := int64(val)
				if target.OverflowInt(x) {
					return fmt.Errorf("overflow error")
				}
				target.SetInt(x)
				return nil
			}
			return fmt.Errorf("value %v cannot be cast to int", value)
		}
	case reflect.String:
		{
			if str, ok := value.(string); ok {
				target.SetString(str)
				return nil
			}
			return fmt.Errorf("value %v cannot be cast to string", value)
		}
	case reflect.Slice:
		return nil
	case reflect.Map:
		return nil
	default:
		return fmt.Errorf("unexpected property type: %v", target.Kind())
	}
}
