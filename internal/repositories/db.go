package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"igo-repo/internal/config"
	"net"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type DatabaseRepository struct {
	ConnectionPool *sql.DB
	schemaName     string
}

type ConnectionProperties struct {
	Host     string
	Port     int
	Database string
	Schema   string
	User     string
	Password string
}

func checkDefined(value string, name string) {
	if value == "" {
		log.Fatalf("Connection property %s undefined", name)
	}
}

func CreateConnectionProperties(options config.Options) ConnectionProperties {
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
		Schema:   options.DBSchemaName,
		User:     options.DBUser,
		Password: options.DBPassword,
	}
}

var errMaybeTransient = errors.New("worth to retry for some time")

func (repo DatabaseRepository) createSchema() error {
	var err error
	var tx *sql.Tx
	tx, err = repo.ConnectionPool.Begin()
	if err != nil {
		if maybeTransient(err) {
			return errMaybeTransient
		}
		return fmt.Errorf("failed to create Tx for schema creation: %w", err)
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow("SELECT count(*) FROM information_schema.schemata where schema_name = $1", repo.schemaName).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existence of schema %s: %w", repo.schemaName, err)
	}

	if count == 1 {
		log.Infof("schema %s already exists", repo.schemaName)
		return nil
	}

	log.Infof("creating schema %s...", repo.schemaName)

	_, err = tx.Exec("create schema " + repo.schemaName)
	if err != nil {
		return fmt.Errorf("failed to create schema: " + repo.schemaName)
	}

	tx.Commit()
	return nil
}

func openConnection(connProps ConnectionProperties) (DatabaseRepository, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&options=-csearch_path=%s",
		connProps.User,
		connProps.Password,
		connProps.Host,
		connProps.Database,
		connProps.Schema,
	)

	db, err := sql.Open("pgx", connStr)
	repo := DatabaseRepository{db, connProps.Schema}
	if err != nil {
		return repo, err
	}
	db.Ping()
	return repo, err
}

func NewDBRepo(connectionProperties ConnectionProperties) (*DatabaseRepository, error) {
	var err error

	repo, errDBOpen := openConnection(connectionProperties)
	if errDBOpen != nil {
		return &repo, errDBOpen
	}

	for i := 0; i < 30; i++ {
		err = repo.createSchema()
		if err == nil {
			return &repo, nil
		} else {
			log.Infof("Failed to create new DatabaseRepository %v; retry count: %v", err, i)
		}
		if errors.Is(err, errMaybeTransient) {
			time.Sleep(2 * time.Second)
			repo.createSchema()
		} else {
			return &repo, fmt.Errorf("create schema failed: %w", err)
		}
	}
	return &repo, err
}

func (repo DatabaseRepository) Close() error {
	return repo.ConnectionPool.Close()
}

func maybeTransient(err error) bool {
	logger := log.WithField("prefix", "maybeTransient")
	logger.Infof("checking db error: %+v.(%T)...", err, err)

	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		logger.Info("Timeout")
		return true
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "dial") && strings.Contains(errMsg, "connect: connection refused") {
		return true
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			logger.Info("unknown host / connection refused")
			return true
		} else if t.Op == "read" {
			logger.Info("connection refused")
			return true
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			logger.Info("connection refused")
			return true
		}
	}

	logger.Debug("non-transient")
	return false
}

func InitDBRepo(configuration config.Options) (*DatabaseRepository, error) {
	logger := log.WithField("prefix", "repositories.InitDBRepo")
	connProps := CreateConnectionProperties(configuration)
	dbRepo, errNewDB := NewDBRepo(connProps)
	if errNewDB != nil {
		logger.Errorf("Failed to create schema %v", errNewDB)
		panic(errNewDB)
	}
	return dbRepo, dbRepo.ExecuteSchemaUpgrade()
}
