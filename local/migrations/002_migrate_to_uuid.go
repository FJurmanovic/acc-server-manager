package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gorm.io/gorm"
)

// Migration002MigrateToUUID migrates tables from integer IDs to UUIDs
type Migration002MigrateToUUID struct {
	DB *gorm.DB
}

// NewMigration002MigrateToUUID creates a new UUID migration
func NewMigration002MigrateToUUID(db *gorm.DB) *Migration002MigrateToUUID {
	return &Migration002MigrateToUUID{DB: db}
}

// Up executes the migration
func (m *Migration002MigrateToUUID) Up() error {
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
func (m *Migration002MigrateToUUID) needsMigration() bool {
	// Check if servers table exists and has integer primary key
	var result struct {
		Type string `gorm:"column:type"`
	}

	err := m.DB.Raw(`
		SELECT type FROM pragma_table_info('servers')
		WHERE name = 'id' AND pk = 1
	`).Scan(&result).Error

	if err != nil || result.Type == "" {
		// Table doesn't exist or no primary key found - assume no migration needed
		return false
	}

	// If the primary key is INTEGER, we need migration
	// If it's TEXT (UUID), migration already done
	return result.Type == "INTEGER" || result.Type == "integer"
}

// Down reverses the migration (not implemented for safety)
func (m *Migration002MigrateToUUID) Down() error {
	logging.Error("UUID migration rollback is not supported for data safety reasons")
	return fmt.Errorf("UUID migration rollback is not supported")
}

// runUUIDMigrationSQL executes the UUID migration using the SQL file
func runUUIDMigrationSQL(db *gorm.DB) error {
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

	// Read the migration SQL from file
	sqlPath := filepath.Join("scripts", "migrations", "002_migrate_servers_to_uuid.sql")
	migrationSQL, err := ioutil.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read migration SQL file: %v", err)
	}

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

// RunUUIDMigration is a convenience function to run the migration
func RunUUIDMigration(db *gorm.DB) error {
	migration := NewMigration002MigrateToUUID(db)
	return migration.Up()
}
