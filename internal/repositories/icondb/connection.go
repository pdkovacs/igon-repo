package icondb

import (
	"database/sql"
	"fmt"
	"igo-repo/internal/config"

	"github.com/rs/zerolog"
)

type connection struct {
	Pool       *sql.DB
	schemaName string
}

func NewDBConnection(conf config.Options, logger zerolog.Logger) (connection, error) {
	sqlDB, errNewDB := open(config.CreateDbProperties(conf, conf.DBSchemaName, logger), logger)
	if errNewDB != nil {
		logger.Error().Msgf("Failed to open connection %v", errNewDB)
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
	logger.Debug().Msgf("connStr=%s", connStr)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	db.Ping()
	return db, nil
}
