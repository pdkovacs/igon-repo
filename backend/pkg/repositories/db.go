package repositories

import (
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	log "github.com/sirupsen/logrus"
)

type ConnectionProperties struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func checkDefined(value string, name string) {
	if value == "" {
		log.Fatalf("Connection property %s undefined", name)
	}
}

func createConnectionProperties(options auxiliaries.Options) ConnectionProperties {
	checkDefined(options.DBHost, "DBHost")
	checkDefined(options.DBName, "DBName")
	if options.DBPort == 0 {
		checkDefined("", "DBPort")
	}
	checkDefined(options.DBUser, "DBUser")
	checkDefined(options.DBPassword, "DBPassword")

	return ConnectionProperties{
		Host:     options.DBHost,
		Port:     options.DBPort,
		Database: options.DBName,
		User:     options.DBUser,
		Password: options.DBPassword,
	}
}
