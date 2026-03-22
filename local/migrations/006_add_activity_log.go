package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"

	"gorm.io/gorm"
)

func RunAddActivityLogMigration(db *gorm.DB) error {
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to ensure migration_records table: %v", err)
	}

	var record MigrationRecord
	if err := db.Where("migration_name = ?", "006_add_activity_log").First(&record).Error; err == nil {
		logging.Info("Activity log migration already applied, skipping")
		return nil
	}

	record = MigrationRecord{
		MigrationName: "006_add_activity_log",
		AppliedAt:     "datetime('now')",
		Success:       true,
		Notes:         "Added ActivityLog table via GORM AutoMigrate",
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to record migration 006: %v", err)
	}

	logging.Info("Activity log migration recorded")
	return nil
}
