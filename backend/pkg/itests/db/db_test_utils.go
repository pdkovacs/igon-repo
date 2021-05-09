package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pdkovacs/igo-repo/backend/pkg/auxiliaries"
	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
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

func getIconCount() (int, error) {
	var getIconCountSQL = "SELECT count(*) from icon"
	var count int
	err := db.QueryRow(getIconCountSQL).Scan(&count)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func createTestIconfile(name, format, size string) domain.Iconfile {
	return domain.Iconfile{
		IconAttributes: domain.IconAttributes{
			Name: name,
		},
		IconfileData: domain.IconfileData{
			IconfileDescriptor: domain.IconfileDescriptor{
				Format: format,
				Size:   size,
			},
			Content: randomBytes(4096),
		},
	}
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
