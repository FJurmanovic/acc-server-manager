package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Get database path from command line args or use default
	dbPath := "acc.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	fmt.Printf("Running UUID migration on database: %s\n", dbPath)

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbPath)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
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

	// Check if migration is needed
	if !needsMigration(db) {
		fmt.Println("Migration not needed - database already uses UUID primary keys")
		return
	}

	fmt.Println("Starting UUID migration...")

	// Run the migration
	if err := runUUIDMigration(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("UUID migration completed successfully!")
}

// needsMigration checks if the UUID migration is needed
func needsMigration(db *gorm.DB) bool {
	// Check if servers table exists and has integer primary key
	var result struct {
		Type string `gorm:"column:type"`
	}

	err := db.Raw(`
		SELECT type FROM pragma_table_info('servers')
		WHERE name = 'id' AND pk = 1
	`).Scan(&result).Error

	if err != nil || result.Type == "" {
		// Table doesn't exist or no primary key found
		return false
	}

	// If the primary key is INTEGER, we need migration
	// If it's TEXT (UUID), migration already done
	return result.Type == "INTEGER" || result.Type == "integer"
}

// runUUIDMigration executes the UUID migration
func runUUIDMigration(db *gorm.DB) error {
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
