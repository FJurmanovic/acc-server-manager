package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

var sqlLineCommentRegexp = regexp.MustCompile(`--[^\n]*`)

type Migration004FixDDLComments struct {
	DB *gorm.DB
}

func NewMigration004FixDDLComments(db *gorm.DB) *Migration004FixDDLComments {
	return &Migration004FixDDLComments{DB: db}
}

func (m *Migration004FixDDLComments) Up() error {
	if err := m.DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %v", err)
	}

	var record MigrationRecord
	if err := m.DB.Where("migration_name = ?", "004_fix_ddl_comments").First(&record).Error; err == nil {
		logging.Info("DDL comment fix migration already applied, skipping")
		return nil
	}

	var tables []string
	if err := m.DB.Raw(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND sql LIKE '%--%' ORDER BY name",
	).Scan(&tables).Error; err != nil {
		return fmt.Errorf("failed to query tables with DDL comments: %v", err)
	}

	if len(tables) == 0 {
		logging.Info("No tables with DDL comments found, skipping")
	}

	for _, table := range tables {
		if err := m.rebuildTable(table); err != nil {
			return fmt.Errorf("failed to rebuild table %q: %v", table, err)
		}
	}

	record = MigrationRecord{
		MigrationName: "004_fix_ddl_comments",
		AppliedAt:     "datetime('now')",
		Success:       true,
		Notes:         fmt.Sprintf("Rebuilt %d table(s): %s", len(tables), strings.Join(tables, ", ")),
	}
	if err := m.DB.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	return nil
}

func (m *Migration004FixDDLComments) rebuildTable(table string) error {
	var originalDDL string
	if err := m.DB.Raw(
		"SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?", table,
	).Row().Scan(&originalDDL); err != nil {
		return err
	}

	cleanDDL := sqlLineCommentRegexp.ReplaceAllString(originalDDL, "")

	tmpTable := table + "__fixed"
	createTmp := strings.Replace(cleanDDL, `"`+table+`"`, "`"+tmpTable+"`", 1)
	if createTmp == cleanDDL {
		createTmp = strings.Replace(cleanDDL, table, "`"+tmpTable+"`", 1)
	}

	logging.Info("Rebuilding table %q to remove DDL comments...", table)

	return m.DB.Transaction(func(tx *gorm.DB) error {
		for _, sql := range []string{
			"PRAGMA foreign_keys = OFF",
			fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tmpTable),
			createTmp,
			fmt.Sprintf("INSERT INTO `%s` SELECT * FROM `%s`", tmpTable, table),
			fmt.Sprintf("DROP TABLE `%s`", table),
			fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`", tmpTable, table),
			"PRAGMA foreign_keys = ON",
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("exec %q: %w", sql, err)
			}
		}
		logging.Info("Table %q rebuilt successfully", table)
		return nil
	})
}

func (m *Migration004FixDDLComments) Down() error {
	return fmt.Errorf("DDL comment fix migration rollback is not supported")
}

func RunFixDDLCommentsMigration(db *gorm.DB) error {
	return NewMigration004FixDDLComments(db).Up()
}
