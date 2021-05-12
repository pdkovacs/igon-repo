package repositories

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func colDefToSQL(columnsDef columnsDefinition, columnName string) string {
	return fmt.Sprintf("%s %s", columnName, columnsDef[columnName])
}

func columnsDefinitionToSQL(columnsDef columnsDefinition) string {
	resultColsDef := ""
	for columnName := range columnsDef {
		if len(resultColsDef) > 0 {
			resultColsDef += ",\n    "
		}
		resultColsDef += colDefToSQL(columnsDef, columnName)
	}
	return resultColsDef
}

func colConstraintsToSQL(colConstraints []string) string {
	if len(colConstraints) == 0 {
		return ""
	}
	return ",\n    " + strings.Join(colConstraints[:], ",\n    ")
}

func makeCreateTableStatement(tableDefinition tableSpec) string {
	format := `CREATE TABLE %s (
	%s%s
)`
	return fmt.Sprintf(
		format,
		tableDefinition.tableName,
		columnsDefinitionToSQL(tableDefinition.columns),
		colConstraintsToSQL(tableDefinition.col_constraints),
	)
}

func createTable(db *sql.DB, tableDefinition tableSpec) error {
	_, err := db.Exec(makeCreateTableStatement(tableDefinition))
	if err != nil {
		return err
	}
	return nil
}

func dropTableIfExists(db *sql.DB, tableDefinition tableSpec) error {
	statement := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableDefinition.tableName)
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	return nil
}

func dropCreateTable(db *sql.DB, tableDefinition tableSpec) error {
	var err error
	err = dropTableIfExists(db, tableDefinition)
	if err != nil {
		return err
	}
	err = createTable(db, tableDefinition)
	if err != nil {
		return err
	}
	return nil
}

func createSchema(db *sql.DB) error {
	var err error
	err = dropCreateTable(db, iconTableSpec)
	if err != nil {
		return err
	}
	err = dropCreateTable(db, iconfileTableSpec)
	if err != nil {
		return err
	}
	return nil
}

func CreateSchemaRetry(db *sql.DB) error {
	var err error

	for i := 0; i < 30; i++ {
		err = createSchema(db)
		if err == nil {
			return nil
		} else {
			log.Infof("CreateSchemaRetry error %v; retry count: %v", err, i)
		}
		if worthIt := isErrorWorthRetrying(err); worthIt {
			time.Sleep(2 * time.Second)
			createSchema(db)
		} else {
			return fmt.Errorf("create schema failed: %w", err)
		}
	}
	return err
}

func isErrorWorthRetrying(err error) bool {
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
