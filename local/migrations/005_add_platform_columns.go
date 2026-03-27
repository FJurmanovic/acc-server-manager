package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"

	"gorm.io/gorm"
)

func RunAddPlatformColumnsMigration(db *gorm.DB) error {
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to ensure migration_records table: %v", err)
	}

	var record MigrationRecord
	if err := db.Where("migration_name = ?", "005_add_platform_columns").First(&record).Error; err == nil {
		logging.Info("Platform columns migration already applied, skipping")
		return nil
	}

	record = MigrationRecord{
		MigrationName: "005_add_platform_columns",
		AppliedAt:     "datetime('now')",
		Success:       true,
		Notes:         "Added Platform and ContainerID columns to servers table via GORM AutoMigrate",
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to record migration 005: %v", err)
	}

	logging.Info("Platform columns migration recorded")
	return nil
}
