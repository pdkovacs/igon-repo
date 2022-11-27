package config

import (
	"fmt"

	"github.com/rs/zerolog"
)

type DbConnectionProperties struct {
	Host     string
	Port     int
	Database string
	Schema   string
	User     string
	Password string
}

func CreateDbProperties(options Options, dbSchema string, logger zerolog.Logger) DbConnectionProperties {
	checkDefined := func(value string, name string) {
		if value == "" {
			msg := fmt.Sprintf("Connection property %s undefined", name)
			logger.Error().Msgf(msg)
			panic(msg)
		}
	}

	checkDefined(options.DBHost, "DBHost")
	checkDefined(options.DBName, "DBName")
	if options.DBPort == 0 {
		checkDefined("", "DBPort")
	}
	checkDefined(options.DBUser, "DBUser")
	checkDefined(options.DBPassword, "DBPassword")

	return DbConnectionProperties{
		Host:     options.DBHost,
		Port:     options.DBPort,
		Database: options.DBName,
		Schema:   dbSchema,
		User:     options.DBUser,
		Password: options.DBPassword,
	}
}
