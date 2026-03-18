package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gorm.io/gorm"
)

type Migration002MigrateToUUID struct {
	DB *gorm.DB
}

func NewMigration002MigrateToUUID(db *gorm.DB) *Migration002MigrateToUUID {
	return &Migration002MigrateToUUID{DB: db}
}

func (m *Migration002MigrateToUUID) Up() error {
	logging.Info("Checking UUID migration...")

	if !m.needsMigration() {
		logging.Info("UUID migration not needed - tables already use UUID primary keys")
		return nil
	}

	logging.Info("Starting UUID migration...")

	var migrationRecord MigrationRecord
	err := m.DB.Where("migration_name = ?", "002_migrate_to_uuid").First(&migrationRecord).Error
	if err == nil {
		logging.Info("UUID migration already applied, skipping")
		return nil
	}

	if err := m.DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %v", err)
	}

	logging.Info("Executing UUID migration...")
	if err := runUUIDMigrationSQL(m.DB); err != nil {
		return fmt.Errorf("failed to execute UUID migration: %v", err)
	}

	logging.Info("UUID migration completed successfully")
	return nil
}

func (m *Migration002MigrateToUUID) needsMigration() bool {
	var result struct {
		Type string `gorm:"column:type"`
	}

	err := m.DB.Raw(`
		SELECT type FROM pragma_table_info('servers')
		WHERE name = 'id' AND pk = 1
	`).Scan(&result).Error

	if err != nil || result.Type == "" {
		return false
	}

	return result.Type == "INTEGER" || result.Type == "integer"
}

func (m *Migration002MigrateToUUID) Down() error {
	logging.Error("UUID migration rollback is not supported for data safety reasons")
	return fmt.Errorf("UUID migration rollback is not supported")
}

func runUUIDMigrationSQL(db *gorm.DB) error {
	if err := db.Exec("PRAGMA foreign_keys=OFF").Error; err != nil {
		return fmt.Errorf("failed to disable foreign keys: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	sqlPath := filepath.Join("scripts", "migrations", "002_migrate_servers_to_uuid.sql")
	migrationSQL, err := ioutil.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read migration SQL file: %v", err)
	}

	if err := tx.Exec(string(migrationSQL)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		return fmt.Errorf("failed to re-enable foreign keys: %v", err)
	}

	return nil
}

func RunUUIDMigration(db *gorm.DB) error {
	migration := NewMigration002MigrateToUUID(db)
	return migration.Up()
}
