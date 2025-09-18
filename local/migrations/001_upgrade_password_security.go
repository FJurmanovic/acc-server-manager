package migrations

import (
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/password"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Migration001UpgradePasswordSecurity struct {
	DB *gorm.DB
}

func NewMigration001UpgradePasswordSecurity(db *gorm.DB) *Migration001UpgradePasswordSecurity {
	return &Migration001UpgradePasswordSecurity{DB: db}
}

func (m *Migration001UpgradePasswordSecurity) Up() error {
	logging.Info("Starting password security upgrade migration...")

	var migrationRecord MigrationRecord
	err := m.DB.Where("migration_name = ?", "001_upgrade_password_security").First(&migrationRecord).Error
	if err == nil {
		logging.Info("Password security migration already applied, skipping")
		return nil
	}

	if err := m.DB.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %v", err)
	}

	tx := m.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec("ALTER TABLE users ADD COLUMN password_backup TEXT").Error; err != nil {
		if !isDuplicateColumnError(err) {
			tx.Rollback()
			return fmt.Errorf("failed to add backup column: %v", err)
		}
	}

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
			continue
		}
		migratedCount++
	}

	if err := tx.Exec("ALTER TABLE users DROP COLUMN password_backup").Error; err != nil {
		logging.Error("Failed to remove backup column (non-critical): %v", err)
	}

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

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration: %v", err)
	}

	logging.Info("Password security migration completed successfully. Migrated: %d, Failed: %d", migratedCount, failedCount)

	if failedCount > 0 {
		logging.Error("Some users failed to migrate. They will need to reset their passwords.")
	}

	return nil
}

func (m *Migration001UpgradePasswordSecurity) migrateUserPassword(tx *gorm.DB, user *UserForMigration) error {
	if isAlreadyHashed(user.Password) {
		logging.Debug("User %s already has hashed password, skipping", user.Username)
		return nil
	}

	if err := tx.Model(user).Update("password_backup", user.Password).Error; err != nil {
		return fmt.Errorf("failed to backup password: %v", err)
	}

	var plainPassword string

	decrypted, err := decryptOldPassword(user.Password)
	if err != nil {
		logging.Error("Failed to decrypt password for user %s, treating as plain text: %v", user.Username, err)

		plainPassword = user.Password

		if len(plainPassword) > 100 || containsBinaryData(plainPassword) {
			return fmt.Errorf("password appears to be corrupted encrypted data")
		}
	} else {
		plainPassword = decrypted
	}

	if plainPassword == "" {
		return errors.New("decrypted password is empty")
	}

	if len(plainPassword) < 1 {
		return errors.New("password too short after decryption")
	}

	hashedPassword, err := password.HashPassword(plainPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	if err := tx.Model(user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	logging.Debug("Successfully migrated password for user %s", user.Username)
	return nil
}

type UserForMigration struct {
	ID       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
}

func (UserForMigration) TableName() string {
	return "users"
}

type MigrationRecord struct {
	ID            uint   `gorm:"primaryKey"`
	MigrationName string `gorm:"unique;not null"`
	AppliedAt     string `gorm:"not null"`
	Success       bool   `gorm:"not null"`
	Notes         string
}

func (MigrationRecord) TableName() string {
	return "migration_records"
}

func isAlreadyHashed(password string) bool {
	return len(password) >= 60 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2y$")
}

func containsBinaryData(s string) bool {
	for _, b := range []byte(s) {
		if b < 32 && b != 9 && b != 10 && b != 13 {
			return true
		}
	}
	return false
}

func isDuplicateColumnError(err error) bool {
	errStr := err.Error()
	return fmt.Sprintf("%v", errStr) == "duplicate column name: password_backup" ||
		fmt.Sprintf("%v", errStr) == "SQLITE_ERROR: duplicate column name: password_backup"
}

func decryptOldPassword(encryptedPassword string) (string, error) {
	return "", errors.New("old decryption not implemented - treating as plain text")
}

func (m *Migration001UpgradePasswordSecurity) Down() error {
	logging.Error("Password security migration rollback is not supported for security reasons")
	return errors.New("password security migration rollback is not supported")
}

func RunPasswordSecurityMigration(db *gorm.DB) error {
	migration := NewMigration001UpgradePasswordSecurity(db)
	return migration.Up()
}
