package repositories

import (
	"database/sql"
	"fmt"
	"igo-repo/internal/config"
	"igo-repo/internal/logging"
	"sort"
	"strings"

	"github.com/rs/zerolog"
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
	conn   dbConnection
	logger zerolog.Logger
}

func OpenDBSchema(config config.Options, dbConn dbConnection, logger zerolog.Logger) (didExist bool, errUpgrade error) {
	schema := dbSchema{
		conn:   dbConn,
		logger: logger,
	}
	didExist, errUpgrade = schema.upgradeSchema()
	if errUpgrade != nil {
		return false, fmt.Errorf("failed to open schema: %w", errUpgrade)
	}
	return
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
	logger := logging.CreateMethodLogger(schema.logger, "executeUpgrade")

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
			logger.Info().Msgf("Version already applied: %s", upgrStep.version)
		} else {
			logger.Info().Msgf("Applying upgrade: '%s' ...", upgrStep.version)
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
	logger := logging.CreateMethodLogger(schema.logger, "createMaybe")
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
		logger.Info().Msgf("schema '%s' exists", schemaName)
		return true, nil
	}

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

func (schema *dbSchema) upgradeSchema() (bool, error) {
	exists, errDoesExist := schema.doesExist()
	if errDoesExist != nil {
		return false, errDoesExist
	}

	if !exists {
		errCreateSchema := schema.create()
		if errCreateSchema != nil {
			return false, errCreateSchema
		}
	}

	if !exists {
		errUpgrade := schema.executeUpgrade()
		if errUpgrade != nil {
			return false, errUpgrade
		}
	}

	return exists, nil
}
