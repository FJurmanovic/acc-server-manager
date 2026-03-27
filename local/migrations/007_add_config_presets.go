package migrations

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ConfigPresetMigration struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"not null"`
	Description   string
	Configuration string `gorm:"type:text"`
	AssistRules   string `gorm:"type:text"`
	Event         string `gorm:"type:text"`
	EventRules    string `gorm:"type:text"`
	Settings      string `gorm:"type:text"`
	CreatedAt     string
	UpdatedAt     string
}

func (ConfigPresetMigration) TableName() string {
	return "config_presets"
}

func RunAddConfigPresetsMigration(db *gorm.DB) error {
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to ensure migration_records table: %v", err)
	}

	var record MigrationRecord
	if err := db.Where("migration_name = ?", "007_add_config_presets").First(&record).Error; err == nil {
		logging.Info("Config presets migration already applied, skipping")
		return nil
	}

	record = MigrationRecord{
		MigrationName: "007_add_config_presets",
		AppliedAt:     time.Now().UTC().Format(time.RFC3339),
		Success:       true,
		Notes:         "Added ConfigPreset table via GORM AutoMigrate",
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to record migration 007: %v", err)
	}

	logging.Info("Config presets migration recorded")
	return nil
}
