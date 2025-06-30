package main

import (
	"acc-server-manager/local/migrations"
	"acc-server-manager/local/utl/logging"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Initialize logging
	logging.Init(true) // Enable debug logging

	// Get database path from command line args or use default
	dbPath := "acc.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	// Make sure we're running from the correct directory
	if !fileExists(dbPath) {
		// Try to find the database in common locations
		possiblePaths := []string{
			"acc.db",
			"../acc.db",
			"../../acc.db",
			"cmd/api/acc.db",
			"../cmd/api/acc.db",
		}

		found := false
		for _, path := range possiblePaths {
			if fileExists(path) {
				dbPath = path
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("Database file not found. Please run from the project root or specify the correct path.")
		}
	}

	// Get absolute path for database
	absDbPath, err := filepath.Abs(dbPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for database: %v", err)
	}

	logging.Info("Using database: %s", absDbPath)

	// Open database connection
	db, err := gorm.Open(sqlite.Open(absDbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations in order
	logging.Info("Starting database migrations...")

	// Migration 001: Password security upgrade (if it exists and hasn't run)
	logging.Info("Checking Migration 001: Password Security Upgrade...")
	if err := migrations.RunPasswordSecurityMigration(db); err != nil {
		log.Fatalf("Migration 001 failed: %v", err)
	}

	// Migration 002: UUID migration
	logging.Info("Checking Migration 002: UUID Migration...")
	if err := migrations.RunUUIDMigration(db); err != nil {
		log.Fatalf("Migration 002 failed: %v", err)
	}

	logging.Info("All migrations completed successfully!")

	// Print summary of migration status
	printMigrationStatus(db)
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// printMigrationStatus prints a summary of applied migrations
func printMigrationStatus(db *gorm.DB) {
	logging.Info("Migration Status Summary:")
	logging.Info("========================")

	// Check if migration_records table exists
	var tableExists int
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migration_records'").Scan(&tableExists).Error
	if err != nil || tableExists == 0 {
		logging.Info("No migration tracking table found - this may be a fresh database")
		return
	}

	// Get all migration records
	var records []struct {
		MigrationName string `gorm:"column:migration_name"`
		AppliedAt     string `gorm:"column:applied_at"`
		Success       bool   `gorm:"column:success"`
		Notes         string `gorm:"column:notes"`
	}

	err = db.Table("migration_records").Find(&records).Error
	if err != nil {
		logging.Error("Failed to fetch migration records: %v", err)
		return
	}

	if len(records) == 0 {
		logging.Info("No migrations have been applied yet")
		return
	}

	for _, record := range records {
		status := "✓ SUCCESS"
		if !record.Success {
			status = "✗ FAILED"
		}

		logging.Info("  %s - %s (%s)", record.MigrationName, status, record.AppliedAt)
		if record.Notes != "" {
			logging.Info("    Notes: %s", record.Notes)
		}
	}

	fmt.Printf("\nTotal migrations applied: %d\n", len(records))
}
