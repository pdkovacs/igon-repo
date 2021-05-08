package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/repositories"
)

var db *sql.DB

func createTestDBPool() {
	connProps := repositories.CreateConnectionProperties(auxiliaries.GetDefaultConfiguration())
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", connProps.User, connProps.Password, connProps.Host, connProps.Database)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	db.Ping()
}

func terminatePool() {
	db.Close()
}

func makeSureHasUptodateDBSchemaWithNoData() {
	err := repositories.CreateSchemaRetry(db)
	if err != nil {
		panic(err)
	}
}

func manageTestResourcesBeforeAll() {
	createTestDBPool()
}

func manageTestResourcesAfterAll() {
	terminatePool()
}

func manageTestResourcesBeforeEach() {
	makeSureHasUptodateDBSchemaWithNoData()
}

func getPool() *sql.DB {
	return db
}
