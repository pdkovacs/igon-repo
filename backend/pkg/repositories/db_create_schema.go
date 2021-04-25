package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
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
	return ",\n    " + fmt.Sprintf(strings.Join(colConstraints[:], ",\n    "))
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

func createSchemaRetry(db *sql.DB) error {
	var err error
	for i := 0; i < 30; i++ {
		err = createSchema(db)
		if err == nil {
			return nil
		}
		if err, ok := err.(*pq.Error); ok {
			error_code_name := err.Code.Name()
			log.Debug("pq error", error_code_name)
			if error_code_name == "sqlclient_unable_to_establish_sqlconnection" ||
				error_code_name == "sqlserver_rejected_establishment_of_sqlconnection" {
				time.Sleep(2 * time.Second)
			} else {
				return err
			}
		} else {
			return err
		}
	}
	return err
}
