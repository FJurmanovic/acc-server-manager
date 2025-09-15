package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"

	"gorm.io/gorm"
)

// UpdateStateHistorySessions migrates tables from integer IDs to UUIDs
type UpdateStateHistorySessions struct {
	DB *gorm.DB
}

// NewUpdateStateHistorySessions creates a new UUID migration
func NewUpdateStateHistorySessions(db *gorm.DB) *UpdateStateHistorySessions {
	return &UpdateStateHistorySessions{DB: db}
}

// Up executes the migration
func (m *UpdateStateHistorySessions) Up() error {
	logging.Info("Checking UUID migration...")

	// Check if migration is needed by looking at the servers table structure
	if !m.needsMigration() {
		logging.Info("UUID migration not needed - tables already use UUID primary keys")
		return nil
	}

	logging.Info("Starting UUID migration...")

	// Check if migration has already been applied
	var migrationRecord MigrationRecord
	err := m.DB.Where("migration_name = ?", "002_migrate_to_uuid").First(&migrationRecord).Error
	if err == nil {
		logging.Info("UUID migration already applied, skipping")
		return nil
	}

	// Create migration tracking table if it doesn't exist
	if err := m.DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %v", err)
	}

	// Execute the UUID migration using the existing migration function
	logging.Info("Executing UUID migration...")
	if err := runUUIDMigrationSQL(m.DB); err != nil {
		return fmt.Errorf("failed to execute UUID migration: %v", err)
	}

	logging.Info("UUID migration completed successfully")
	return nil
}

// needsMigration checks if the UUID migration is needed by examining table structure
func (m *UpdateStateHistorySessions) needsMigration() bool {
	// Check if servers table exists and has integer primary key
	var result struct {
		Exists bool `gorm:"column:exists"`
	}

	err := m.DB.Raw(`
		SELECT count(*) > 0 as exists FROM state_history
		WHERE length(session) > 1 LIMIT 1;
	`).Scan(&result).Error

	if err != nil || !result.Exists {
		// Table doesn't exist or no primary key found - assume no migration needed
		return false
	}
	return result.Exists
}

// Down reverses the migration (not implemented for safety)
func (m *UpdateStateHistorySessions) Down() error {
	logging.Error("UUID migration rollback is not supported for data safety reasons")
	return fmt.Errorf("UUID migration rollback is not supported")
}

// runUpdateStateHistorySessionsMigration executes the UUID migration using the SQL file
func runUpdateStateHistorySessionsMigration(db *gorm.DB) error {
	// Disable foreign key constraints during migration
	if err := db.Exec("PRAGMA foreign_keys=OFF").Error; err != nil {
		return fmt.Errorf("failed to disable foreign keys: %v", err)
	}

	// Start transaction
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

	// Execute the migration
	if err := tx.Exec(string(migrationSQL)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration: %v", err)
	}

	// Re-enable foreign key constraints
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		return fmt.Errorf("failed to re-enable foreign keys: %v", err)
	}

	return nil
}

// RunUpdateStateHistorySessionsMigration is a convenience function to run the migration
func RunUpdateStateHistorySessionsMigration(db *gorm.DB) error {
	migration := NewUpdateStateHistorySessions(db)
	return migration.Up()
}
