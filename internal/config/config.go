package config

import (
	"encoding/json"
	"fmt"
	"iconrepo/internal/app/security/authn"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/rs/zerolog/log"
)

// PasswordCredentials holds password-credentials
type PasswordCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UsersByRoles maps roles to lists of user holding the role
type UsersByRoles map[string][]string

// Options holds the available command-line options
type Options struct {
	ServerHostname              string                     `json:"serverHostname" env:"SERVER_HOSTNAME" long:"server-hostname" short:"h" default:"localhost" description:"Server hostname"`
	ServerPort                  int                        `json:"serverPort" env:"SERVER_PORT" long:"server-port" short:"p" default:"8080" description:"Server port"`
	ServerURLContext            string                     `json:"serverUrlContext" env:"SERVER_URL_CONTEXT" long:"server-url-context" short:"c" default:"" description:"Server url context"`
	SessionMaxAge               int                        `json:"sessionMaxAge" env:"SESSION_MAX_AGE" long:"session-max-age" short:"s" default:"86400" description:"The maximum age in secods of a user's session"`
	LoadBalancerAddress         string                     `json:"loadBalancerAddress" env:"LOAD_BALANCER_ADDRESS" long:"load-balancer-address" short:"" default:"" description:"The load balancer address patter"`
	AppDescription              string                     `json:"appDescription" env:"APP_DESCRIPTION" long:"app-description" short:"" default:"" description:"Application description"`
	SessionDbName               string                     `json:"sessionDbName" env:"SESSION_DB_NAME" long:"session-db-name" short:"" default:"" description:"Name of the session DB"`
	LocalGitRepo                string                     `json:"localGitRepo" env:"LOCAL_GIT_REPO" long:"local-git-repo" short:"g" default:"" description:"Path to the local git repository"`
	GitlabNamespacePath         string                     `json:"gitlabNamespacePath" env:"GITLAB_NAMESPACE_PATH" long:"gitlab-namespace-path" short:"" default:"" description:"GitLab namespace path"`
	GitlabProjectPath           string                     `json:"gitlabProjectPath" env:"GITLAB_PROJECT_PATH" long:"gitlab-project-path" short:"" default:"icon-repo-gitrepo-test" description:"GitLab project path"`
	GitlabMainBranch            string                     `json:"gitlabMainBranch" env:"GITLAB_MAIN_BRANCH" long:"gitlab-main-branch" short:"" default:"main" description:"The GitLab project's main branch"`
	GitlabAccessToken           string                     `json:"gitlabAccessToken" env:"GITLAB_ACCESS_TOKEN" long:"gitlab-access-token" short:"" default:"" description:"GitLab API access token"`
	AuthenticationType          authn.AuthenticationScheme `json:"authenticationType" env:"AUTHENTICATION_TYPE" long:"authentication-type" short:"a" default:"oidc" description:"Authentication type"`
	PasswordCredentials         []PasswordCredentials      `json:"passwordCredentials" env:"PASSWORD_CREDENTIALS" long:"password-credentials"`
	OIDCClientID                string                     `json:"oidcClientId" env:"OIDC_CLIENT_ID" long:"oidc-client-id" short:"" default:"" description:"OIDC client id"`
	OIDCClientSecret            string                     `json:"oidcClientSecret" env:"OIDC_CLIENT_SECRET" long:"oidc-client-secret" short:"" default:"" description:"OIDC client secret"`
	OIDCAccessTokenURL          string                     `json:"oidcAccessTokenUrl" env:"OIDC_ACCESS_TOKEN_URL" long:"oidc-access-token-url" short:"" default:"" description:"OIDC access token url"`
	OIDCUserAuthorizationURL    string                     `json:"oidcUserAuthorizationUrl" env:"OIDC_USER_AUTHORIZATION_URL" long:"oidc-user-authorization-url" short:"" default:"" description:"OIDC user authorization url"`
	OIDCClientRedirectBackURL   string                     `json:"oidcClientRedirectBackUrl" env:"OIDC_CLIENT_REDIRECT_BACK_URL" long:"oidc-client-redirect-back-url" short:"" default:"" description:"OIDC client redirect back url"`
	OIDCTokenIssuer             string                     `json:"oidcTokenIssuer" env:"OIDC_TOKEN_ISSUER" long:"oidc-token-issuer" short:"" default:"" description:"OIDC token issuer"`
	OIDCLogoutURL               string                     `json:"oidcLogoutUrl" env:"OIDC_LOGOUT_URL" long:"oidc-logout-url" short:"" default:"" description:"OIDC logout URL"`
	OIDCIpJwtPublicKeyURL       string                     `json:"oidcIpJwtPublicKeyUrl" env:"OIDC_IP_JWT_PUBLIC_KEY_URL" long:"oidc-ip-jwt-public-key-url" short:"" default:"" description:"OIDC ip jwt public key url"`
	OIDCIpJwtPublicKeyPemBase64 string                     `json:"oidcIpJwtPublicKeyPemBase64" env:"OIDC_IP_JWT_PUBLIC_KEY_PEM_BASE64" long:"oidc-ip-jwt-public-key-pem-base64" short:"" default:"" description:"OIDC ip jwt public key pem base64"`
	OIDCIpLogoutURL             string                     `json:"oidcIpLogoutUrl" env:"OIDC_IP_LOGOUT_URL" long:"oidc-ip-logout-url" short:"" default:"" description:"OIDC ip logout url"`
	UsersByRoles                UsersByRoles               `json:"usersByRoles" env:"USERS_BY_ROLES" long:"users-by-roles" short:"" default:"" description:"Users by roles"`
	DBHost                      string                     `json:"dbHost" env:"DB_HOST" long:"db-host" short:"" default:"localhost" description:"DB host"`
	DBPort                      int                        `json:"dbPort" env:"DB_PORT" long:"db-port" short:"" default:"5432" description:"DB port"`
	DBUser                      string                     `json:"dbUser" env:"DB_USER" long:"db-user" short:"" default:"iconrepo" description:"DB user"`
	DBPassword                  string                     `json:"dbPassword" env:"DB_PASSWORD" long:"db-password" short:"" default:"iconrepo" description:"DB password"`
	DBName                      string                     `json:"dbName" env:"DB_NAME" long:"db-name" short:"" default:"iconrepo" description:"Name of the database"`
	DBSchemaName                string                     `json:"dbSchemaName" env:"DB_SCHEMA_NAME" long:"db-schema-name" short:"" default:"icon_repo" description:"Name of the database schemma"`
	EnableBackdoors             bool                       `json:"enableBackdoors" env:"ENABLE_BACKDOORS" long:"enable-backdoors" short:"" description:"Enable backdoors"`
	UsernameCookie              string                     `json:"usernameCookie" env:"USERNAME_COOKIE" long:"username-cookie" short:"" description:"The name of the cookie, if any, carrying username. Only OIDC for now."`
	LogLevel                    string                     `json:"logLevel" env:"LOG_LEVEL" long:"log-level" short:"l" default:"info"`
	AllowedClientURLsRegex      string                     `json:"allowedClientUrlsRegex" env:"ALLOWED_CLIENT_URLS_REGEX" long:"allowed-client-urls-regex" short:"" default:""`
	DynamoDBURL                 string                     `json:"dynamoDbUrl" env:"DYNAMO_DB_URL" long:"dynamo-db-url" short:"" default:""`
}

var DefaultIconRepoHome = filepath.Join(os.Getenv("HOME"), ".ui-toolbox/icon-repo")
var DefaultIconDataLocationGit = filepath.Join(DefaultIconRepoHome, "git-repo")
var DefaultConfigFilePath = filepath.Join(DefaultIconRepoHome, "config.json")

type ConfigFilePath string

const (
	NO_CONFIG_FILE ConfigFilePath = "none"
)

// GetConfigFilePath gets the path of the configuration file
func GetConfigFilePath() ConfigFilePath {
	result := os.Getenv("ICON_REPO_CONFIG_FILE")
	if result == "" {
		result = DefaultConfigFilePath
	}
	log.Info().Str("configfile_path", result).Msg("configuration file found")
	return ConfigFilePath(result)
}

// ReadConfiguration reads the configuration file and merges it with the command line arguments
func ReadConfiguration(filePath ConfigFilePath, clArgs []string) (Options, error) {
	mapInConfigFile := map[string]interface{}{}
	optsInFile := Options{}
	if filePath != NO_CONFIG_FILE {
		var err error
		mapInConfigFile, optsInFile, err = readConfigurationFromFile(filePath)
		if err != nil {
			return Options{}, err
		}
	}
	return parseFlagsMergeSettings(clArgs, mapInConfigFile, optsInFile), nil
}

func readConfigurationFromFile(configFilePath ConfigFilePath) (map[string]interface{}, Options, error) {
	filePath := string(configFilePath)
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

	if opts.LocalGitRepo == "" {
		opts.LocalGitRepo = DefaultIconDataLocationGit
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
	case reflect.Bool:
		{
			var x bool
			if value == "true" {
				x = true
			} else if value == "false" {
				x = false
			} else {
				return fmt.Errorf("expected 'true' or 'false', found: %s", value)
			}
			target.SetBool(x)
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

func UseCORS(options Options) bool {
	return len(options.AllowedClientURLsRegex) > 0
}
