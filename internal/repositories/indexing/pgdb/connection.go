package pgdb

import (
	"database/sql"
	"fmt"
	"iconrepo/internal/config"
	"iconrepo/internal/logging"

	"github.com/rs/zerolog"
)

type connection struct {
	Pool       *sql.DB
	schemaName string
}

func NewDBConnection(conf config.Options) (connection, error) {
	logger := logging.Get()
	sqlDB, errNewDB := open(config.CreateDbProperties(conf, conf.DBSchemaName, logger), logger)
	if errNewDB != nil {
		logger.Error().Err(errNewDB).Msg("failed to open connection")
		panic(errNewDB)
	}
	return connection{sqlDB, conf.DBSchemaName}, nil
}

func open(connProps config.DbConnectionProperties, logger zerolog.Logger) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&options=-csearch_path=%s",
		connProps.User,
		connProps.Password,
		connProps.Host,
		connProps.Port,
		connProps.Database,
		connProps.Schema,
	)
	logger.Debug().Str("connection-string", connStr).Msg("opening connection...")
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		logger.Error().Err(err).Str("user_name", connProps.User).Str("password", connProps.Password).Str("full_connstr", connStr).Msg("failed to open database")
		return nil, err
	}
	db.Ping()
	return db, nil
}
