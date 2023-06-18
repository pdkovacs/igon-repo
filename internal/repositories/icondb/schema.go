package icondb

import (
	"database/sql"
	"fmt"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"sort"
	"strings"
)

type upgradeStep struct {
	version string
	sqls    []string
}

var upgradeSteps = []upgradeStep{
	{
		version: "2018-04-04/0 - first version",
		sqls: []string{
			`CREATE TABLE icon(
				id          serial primary key,
				name        text,
				modified_by text,
				modified_at timestamp DEFAULT now(),
				UNIQUE(name)
			)`,
			`CREATE TABLE icon_file(
				id          serial primary key,
				icon_id     int REFERENCES icon(id) ON DELETE CASCADE,
				file_format text,
				icon_size   text,
				content     bytea,
				UNIQUE (icon_id, file_format, icon_size)
			)`,
		},
	},
	{
		version: "2018-12-30/1 - tag support",
		sqls: []string{
			"CREATE TABLE tag(id serial primary key, text text)",
			"CREATE TABLE icon_to_tags (" +
				"icon_id int REFERENCES icon(id) ON DELETE CASCADE, " +
				"tag_id  int REFERENCES tag(id)  ON DELETE RESTRICT" +
				")",
		},
	},
}

type dbSchema struct {
	conn connection
}

// OpenSchema checks the availability of the schema, creates and upgrades it as necessary
// Returns true if the schema already existed.
func OpenSchema(config config.Options, dbConn connection) (bool, error) {
	schema := dbSchema{
		conn: dbConn,
	}

	schemaExists, schemaExistErr := schema.doesExist()
	if schemaExistErr != nil {
		return false, fmt.Errorf("failed to open database schema: %w", schemaExistErr)
	}

	if !schemaExists {
		errCreateSchema := schema.create()

		if errCreateSchema != nil {
			return false, errCreateSchema
		}
	}

	upgradeErr := schema.executeUpgrade()
	if upgradeErr != nil {
		return false, fmt.Errorf("failed upgrade the schema: %w", upgradeErr)
	}

	return schemaExists, nil
}

func compareVersions(upgrStep1 upgradeStep, upgrStep2 upgradeStep) int {
	return strings.Compare(upgrStep1.version, upgrStep2.version)
}

func makeSureMetaExists(tx *sql.Tx) error {
	const sql = "CREATE TABLE IF NOT EXISTS meta (version TEXT primary key, upgrade_date TIMESTAMP)"
	_, err := tx.Exec(sql)
	if err != nil {
		return fmt.Errorf("failed to make sure that meta table exist: %w", err)
	}
	return nil
}

func isUpgradeApplied(tx *sql.Tx, version string) (bool, error) {
	var err error
	err = makeSureMetaExists(tx)
	if err != nil {
		return false, fmt.Errorf("failed to determine whether schema upgrade %v has been applied or not: %w", version, err)
	}
	const sql = "SELECT count(*) as upgrade_count FROM meta WHERE version = $1"
	var upgradeCount = 0
	err = tx.QueryRow(sql, version).Scan(&upgradeCount)
	if err != nil {
		return false, fmt.Errorf("failed to determine whether schema upgrade %v has been applied or not: %w", version, err)
	}
	return upgradeCount > 0, nil
}

func createMetaRecord(tx *sql.Tx, version string) error {
	_, err := tx.Exec("INSERT INTO meta(version, upgrade_date) VALUES($1, current_timestamp)", version)
	if err != nil {
		return fmt.Errorf("failed to make a record of the schema upgrade to %s: %w", version, err)
	}
	return nil
}

func applyUpgrade(tx *sql.Tx, upgrStep upgradeStep) error {
	var err error
	for _, uStatement := range upgrStep.sqls {
		_, err = tx.Exec(uStatement)
		if err != nil {
			return fmt.Errorf("failed to execute schema upgrade step %s for version %s: %w", uStatement, upgrStep.version, err)
		}
	}
	err = createMetaRecord(tx, upgrStep.version)
	if err != nil {
		return fmt.Errorf("failed to apply schema upgrade to %s: %w", upgrStep.version, err)
	}
	return nil
}

func (schema *dbSchema) executeUpgrade() error {
	var err error

	logger := logging.Get().With().Str(logging.UnitLogger, "db-schema").Str(logging.MethodLogger, "executeUpgarde").Logger()

	sort.Slice(upgradeSteps, func(i int, j int) bool { return compareVersions(upgradeSteps[i], upgradeSteps[j]) < 0 })

	var tx *sql.Tx
	tx, err = schema.conn.Pool.Begin()
	if err != nil {
		return fmt.Errorf("failed to execute schema upgrade: %w", err)
	}
	defer tx.Rollback()

	for _, upgrStep := range upgradeSteps {
		var applied bool
		applied, err = isUpgradeApplied(tx, upgrStep.version)
		if err != nil {
			return fmt.Errorf("failed to execute schema upgrade: %w", err)
		}
		if applied {
			logger.Info().Str("version", upgrStep.version).Msg("already applied version found")
		} else {
			logger.Info().Str("version", upgrStep.version).Msg("Applying upgrade...")
			err = applyUpgrade(tx, upgrStep)
			if err != nil {
				return fmt.Errorf("failed to apply upgrade step '%s': %w", upgrStep.version, err)
			}
		}
	}
	tx.Commit()
	return nil
}

func (schema *dbSchema) doesExist() (bool, error) {
	logger := logging.Get().With().Str(logging.UnitLogger, "db-schema").Str(logging.MethodLogger, "doesExist").Logger()

	db := schema.conn.Pool
	schemaName := schema.conn.schemaName
	row := db.QueryRow("SELECT schema_name FROM information_schema.schemata WHERE schema_name = $1", schemaName)
	errQuery := row.Err()
	if errQuery != nil {
		return false, fmt.Errorf("failed to query whether schema '%s' exists or not: %w", schemaName, errQuery)
	}

	var scheName sql.NullString
	errRowScan := row.Scan(&scheName)
	if errRowScan != nil && errRowScan != sql.ErrNoRows {
		return false, errRowScan
	}

	if scheName.Valid && scheName.String == schemaName {
		logger.Info().Str("schema-name", schemaName).Msg("schema exists")
		return true, nil
	}

	logger.Info().Str("schema-name", schemaName).Msg("schema doesn't exist")

	return false, nil
}

func (schema *dbSchema) create() error {
	db := schema.conn.Pool
	schemaName := schema.conn.schemaName
	_, errCreate := db.Exec("create schema " + schemaName)
	if errCreate != nil {
		return fmt.Errorf("failed to create schema '%s': %w", schemaName, errCreate)
	}
	return nil
}

func (schema *dbSchema) delete() error {
	db := schema.conn.Pool
	schemaName := schema.conn.schemaName
	_, errDelete := db.Exec("drop schema " + schemaName + " cascade")
	if errDelete != nil {
		return fmt.Errorf("failed to delete schema '%s': %w", schemaName, errDelete)
	}
	return nil
}
