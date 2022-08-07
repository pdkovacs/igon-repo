package repositories

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
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

func (repo DatabaseRepository) ExecuteSchemaUpgrade() error {
	var err error
	logger := log.WithField("prefix", "execute-schema-upgrade")
	sort.Slice(upgradeSteps, func(i int, j int) bool { return compareVersions(upgradeSteps[i], upgradeSteps[j]) < 0 })

	var tx *sql.Tx
	tx, err = repo.ConnectionPool.Begin()
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
			logger.Infof("Version already applied: %s", upgrStep.version)
		} else {
			logger.Infof("Applying upgrade: '%s' ...", upgrStep.version)
			err = applyUpgrade(tx, upgrStep)
			if err != nil {
				return fmt.Errorf("failed to apply upgrade step '%s': %w", upgrStep.version, err)
			}
		}
	}
	tx.Commit()
	return nil
}
