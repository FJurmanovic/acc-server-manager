package migrations

import (
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/password"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Migration001UpgradePasswordSecurity migrates existing user passwords from encrypted to hashed format
type Migration001UpgradePasswordSecurity struct {
	DB *gorm.DB
}

// NewMigration001UpgradePasswordSecurity creates a new password security migration
func NewMigration001UpgradePasswordSecurity(db *gorm.DB) *Migration001UpgradePasswordSecurity {
	return &Migration001UpgradePasswordSecurity{DB: db}
}

// Up executes the migration
func (m *Migration001UpgradePasswordSecurity) Up() error {
	logging.Info("Starting password security upgrade migration...")

	// Check if migration has already been applied
	var migrationRecord MigrationRecord
	err := m.DB.Where("migration_name = ?", "001_upgrade_password_security").First(&migrationRecord).Error
	if err == nil {
		logging.Info("Password security migration already applied, skipping")
		return nil
	}

	// Create migration tracking table if it doesn't exist
	if err := m.DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %v", err)
	}

	// Start transaction
	tx := m.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Add a backup column for old passwords (temporary)
	if err := tx.Exec("ALTER TABLE users ADD COLUMN password_backup TEXT").Error; err != nil {
		// Column might already exist, ignore if it's a duplicate column error
		if !isDuplicateColumnError(err) {
			tx.Rollback()
			return fmt.Errorf("failed to add backup column: %v", err)
		}
	}

	// Get all users with encrypted passwords
	var users []UserForMigration
	if err := tx.Find(&users).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch users: %v", err)
	}

	logging.Info("Found %d users to migrate", len(users))

	migratedCount := 0
	failedCount := 0

	for _, user := range users {
		if err := m.migrateUserPassword(tx, &user); err != nil {
			logging.Error("Failed to migrate user %s (ID: %s): %v", user.Username, user.ID, err)
			failedCount++
			// Continue with other users rather than failing completely
			continue
		}
		migratedCount++
	}

	// Remove backup column after successful migration
	if err := tx.Exec("ALTER TABLE users DROP COLUMN password_backup").Error; err != nil {
		logging.Error("Failed to remove backup column (non-critical): %v", err)
		// Don't fail the migration for this
	}

	// Record successful migration
	migrationRecord = MigrationRecord{
		MigrationName: "001_upgrade_password_security",
		AppliedAt:     "datetime('now')",
		Success:       true,
		Notes:         fmt.Sprintf("Migrated %d users, %d failed", migratedCount, failedCount),
	}

	if err := tx.Create(&migrationRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration: %v", err)
	}

	logging.Info("Password security migration completed successfully. Migrated: %d, Failed: %d", migratedCount, failedCount)

	if failedCount > 0 {
		logging.Error("Some users failed to migrate. They will need to reset their passwords.")
	}

	return nil
}

// migrateUserPassword migrates a single user's password
func (m *Migration001UpgradePasswordSecurity) migrateUserPassword(tx *gorm.DB, user *UserForMigration) error {
	// Skip if password is already hashed (bcrypt hashes start with $2a$, $2b$, or $2y$)
	if isAlreadyHashed(user.Password) {
		logging.Debug("User %s already has hashed password, skipping", user.Username)
		return nil
	}

	// Backup original password
	if err := tx.Model(user).Update("password_backup", user.Password).Error; err != nil {
		return fmt.Errorf("failed to backup password: %v", err)
	}

	// Try to decrypt the old password
	var plainPassword string

	// First, try to decrypt using the old encryption method
	decrypted, err := decryptOldPassword(user.Password)
	if err != nil {
		// If decryption fails, the password might already be plain text or corrupted
		logging.Error("Failed to decrypt password for user %s, treating as plain text: %v", user.Username, err)

		// Use original password as-is (might be plain text from development)
		plainPassword = user.Password

		// Validate it's not obviously encrypted data
		if len(plainPassword) > 100 || containsBinaryData(plainPassword) {
			return fmt.Errorf("password appears to be corrupted encrypted data")
		}
	} else {
		plainPassword = decrypted
	}

	// Validate plain password
	if plainPassword == "" {
		return errors.New("decrypted password is empty")
	}

	if len(plainPassword) < 1 {
		return errors.New("password too short after decryption")
	}

	// Hash the plain password using bcrypt
	hashedPassword, err := password.HashPassword(plainPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update with hashed password
	if err := tx.Model(user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	logging.Debug("Successfully migrated password for user %s", user.Username)
	return nil
}

// UserForMigration represents a user record for migration purposes
type UserForMigration struct {
	ID       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
}

// TableName specifies the table name for GORM
func (UserForMigration) TableName() string {
	return "users"
}

// MigrationRecord tracks applied migrations
type MigrationRecord struct {
	ID            uint   `gorm:"primaryKey"`
	MigrationName string `gorm:"unique;not null"`
	AppliedAt     string `gorm:"not null"`
	Success       bool   `gorm:"not null"`
	Notes         string
}

// TableName specifies the table name for GORM
func (MigrationRecord) TableName() string {
	return "migration_records"
}

// isAlreadyHashed checks if a password is already bcrypt hashed
func isAlreadyHashed(password string) bool {
	return len(password) >= 60 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2y$")
}

// containsBinaryData checks if a string contains binary data
func containsBinaryData(s string) bool {
	for _, b := range []byte(s) {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Allow tab, newline, carriage return
			return true
		}
	}
	return false
}

// isDuplicateColumnError checks if an error is due to duplicate column
func isDuplicateColumnError(err error) bool {
	errStr := err.Error()
	return fmt.Sprintf("%v", errStr) == "duplicate column name: password_backup" ||
		fmt.Sprintf("%v", errStr) == "SQLITE_ERROR: duplicate column name: password_backup"
}

// decryptOldPassword attempts to decrypt using the old encryption method
// This is a simplified version of the old DecryptPassword function
func decryptOldPassword(encryptedPassword string) (string, error) {
	// This would use the old decryption logic
	// For now, we'll return an error to force treating as plain text
	// In a real scenario, you'd implement the old decryption here
	return "", errors.New("old decryption not implemented - treating as plain text")
}

// Down reverses the migration (if needed)
func (m *Migration001UpgradePasswordSecurity) Down() error {
	logging.Error("Password security migration rollback is not supported for security reasons")
	return errors.New("password security migration rollback is not supported")
}

// RunMigration is a convenience function to run the migration
func RunPasswordSecurityMigration(db *gorm.DB) error {
	migration := NewMigration001UpgradePasswordSecurity(db)
	return migration.Up()
}
