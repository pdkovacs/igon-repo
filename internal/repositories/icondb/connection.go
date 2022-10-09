package icondb

import (
	"database/sql"
	"fmt"
	"igo-repo/internal/config"

	"github.com/rs/zerolog"
)

type connectionProperties struct {
	host     string
	port     int
	database string
	schema   string
	user     string
	password string
}

type connection struct {
	Pool       *sql.DB
	schemaName string
}

func NewDBConnection(config config.Options, logger zerolog.Logger) (connection, error) {
	sqlDB, errNewDB := open(createProperties(config, config.DBSchemaName, logger), logger)
	if errNewDB != nil {
		logger.Error().Msgf("Failed to open connection %v", errNewDB)
		panic(errNewDB)
	}
	return connection{sqlDB, config.DBSchemaName}, nil
}

func createProperties(options config.Options, dbSchema string, logger zerolog.Logger) connectionProperties {
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

	return connectionProperties{
		host:     options.DBHost,
		port:     options.DBPort,
		database: options.DBName,
		schema:   dbSchema,
		user:     options.DBUser,
		password: options.DBPassword,
	}
}

func open(connProps connectionProperties, logger zerolog.Logger) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&options=-csearch_path=%s",
		connProps.user,
		connProps.password,
		connProps.host,
		connProps.port,
		connProps.database,
		connProps.schema,
	)
	logger.Debug().Msgf("connStr=%s", connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	db.Ping()
	return db, nil
}
