package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"

	"gorm.io/gorm"
)

type UpdateStateHistorySessions struct {
	DB *gorm.DB
}

func NewUpdateStateHistorySessions(db *gorm.DB) *UpdateStateHistorySessions {
	return &UpdateStateHistorySessions{DB: db}
}

func (m *UpdateStateHistorySessions) Up() error {
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

func (m *UpdateStateHistorySessions) needsMigration() bool {
	var result struct {
		Exists bool `gorm:"column:exists"`
	}

	err := m.DB.Raw(`
		SELECT count(*) > 0 as exists FROM state_history
		WHERE length(session) > 1 LIMIT 1;
	`).Scan(&result).Error

	if err != nil || !result.Exists {
		return false
	}
	return result.Exists
}

func (m *UpdateStateHistorySessions) Down() error {
	logging.Error("UUID migration rollback is not supported for data safety reasons")
	return fmt.Errorf("UUID migration rollback is not supported")
}

func runUpdateStateHistorySessionsMigration(db *gorm.DB) error {
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

	migrationSQL := "UPDATE state_history SET session = upper(substr(session, 1, 1));"

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

func RunUpdateStateHistorySessionsMigration(db *gorm.DB) error {
	migration := NewUpdateStateHistorySessions(db)
	return migration.Up()
}
