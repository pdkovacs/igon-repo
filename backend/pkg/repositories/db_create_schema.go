package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var errMaybeTransient = errors.New("Worth to retry for some time")

func createSchema(db *sql.DB, schemaName string) error {
	var err error
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		if maybeTransient(err) {
			return errMaybeTransient
		}
		return fmt.Errorf("failed to create Tx for schema creation: %w", err)
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow("SELECT count(*) FROM information_schema.schemata where schema_name = $1", schemaName).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existence of schema %s: %w", schemaName, err)
	}

	if count == 1 {
		log.Infof("schema %s already exists", schemaName)
		return nil
	}

	log.Infof("creating schema %s...", schemaName)

	_, err = tx.Exec("create schema " + schemaName)
	if err != nil {
		return fmt.Errorf("failed to create schema: " + schemaName)
	}

	tx.Commit()
	return nil
}

func CreateSchemaRetry(db *sql.DB, schemaName string) error {
	var err error

	for i := 0; i < 30; i++ {
		err = createSchema(db, schemaName)
		if err == nil {
			return nil
		} else {
			log.Infof("CreateSchemaRetry error %v; retry count: %v", err, i)
		}
		if errors.Is(err, errMaybeTransient) {
			time.Sleep(2 * time.Second)
			createSchema(db, schemaName)
		} else {
			return fmt.Errorf("create schema failed: %w", err)
		}
	}
	return err
}

func maybeTransient(err error) bool {
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		println("Timeout")
		return true
	}

	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			println("unknown host / connection refused")
			return true
		} else if t.Op == "read" {
			println("connection refused")
			return true
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			println("connection refused")
			return true
		}
	}
	return false
}
