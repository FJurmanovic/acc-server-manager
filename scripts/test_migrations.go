package main

import (
	"acc-server-manager/local/migrations"
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Initialize logging
	logging.Init(true) // Enable debug logging

	// Create a test database
	testDbPath := "test_migrations.db"

	// Remove existing test database if it exists
	if fileExists(testDbPath) {
		os.Remove(testDbPath)
	}

	logging.Info("Creating test database: %s", testDbPath)

	// Open database connection
	db, err := gorm.Open(sqlite.Open(testDbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Create initial schema with integer IDs to simulate old database
	logging.Info("Creating initial schema with integer IDs...")
	createOldSchema(db)

	// Insert test data with integer IDs
	logging.Info("Inserting test data...")
	insertTestData(db)

	// Run UUID migration
	logging.Info("Running UUID migration...")
	if err := migrations.RunUUIDMigration(db); err != nil {
		log.Fatalf("UUID migration failed: %v", err)
	}

	// Verify migration worked
	logging.Info("Verifying migration results...")
	if err := verifyMigration(db); err != nil {
		log.Fatalf("Migration verification failed: %v", err)
	}

	// Test role system
	logging.Info("Testing role system...")
	if err := testRoleSystem(db); err != nil {
		log.Fatalf("Role system test failed: %v", err)
	}

	// Test Super Admin deletion prevention
	logging.Info("Testing Super Admin deletion prevention...")
	if err := testSuperAdminDeletionPrevention(db); err != nil {
		log.Fatalf("Super Admin deletion prevention test failed: %v", err)
	}

	logging.Info("All tests passed successfully!")

	// Clean up
	os.Remove(testDbPath)
	logging.Info("Test database cleaned up")
}

func createOldSchema(db *gorm.DB) {
	// Create tables with integer primary keys to simulate old schema
	db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			ip TEXT NOT NULL,
			port INTEGER NOT NULL,
			config_path TEXT NOT NULL,
			service_name TEXT NOT NULL,
			date_created DATETIME,
			from_steam_cmd BOOLEAN NOT NULL DEFAULT 1
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL,
			config_file TEXT NOT NULL,
			old_config TEXT,
			new_config TEXT,
			changed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS state_histories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL,
			session TEXT,
			track TEXT,
			player_count INTEGER,
			date_created DATETIME,
			session_start DATETIME,
			session_duration_minutes INTEGER,
			session_id INTEGER NOT NULL DEFAULT 0
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS steam_credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			date_created DATETIME,
			last_updated DATETIME
		)
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS system_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT,
			value TEXT,
			default_value TEXT,
			description TEXT,
			date_modified TEXT
		)
	`)
}

func insertTestData(db *gorm.DB) {
	// Insert test server
	db.Exec(`
		INSERT INTO servers (name, ip, port, config_path, service_name, date_created, from_steam_cmd)
		VALUES ('Test Server', '127.0.0.1', 9600, '/test/path', 'TestService', datetime('now'), 1)
	`)

	// Insert test config
	db.Exec(`
		INSERT INTO configs (server_id, config_file, old_config, new_config)
		VALUES (1, 'test.json', '{"old": true}', '{"new": true}')
	`)

	// Insert test state history
	db.Exec(`
		INSERT INTO state_histories (server_id, session, track, player_count, date_created, session_duration_minutes, session_id)
		VALUES (1, 'Practice', 'monza', 5, datetime('now'), 60, 1)
	`)

	// Insert test steam credentials
	db.Exec(`
		INSERT INTO steam_credentials (username, password, date_created, last_updated)
		VALUES ('testuser', 'testpass', datetime('now'), datetime('now'))
	`)

	// Insert test system config
	db.Exec(`
		INSERT INTO system_configs (key, value, default_value, description, date_modified)
		VALUES ('test_key', 'test_value', 'default_value', 'Test config', datetime('now'))
	`)
}

func verifyMigration(db *gorm.DB) error {
	// Check that all tables now have UUID primary keys

	// Check servers table
	var serverID string
	err := db.Raw("SELECT id FROM servers LIMIT 1").Scan(&serverID).Error
	if err != nil {
		return fmt.Errorf("failed to query servers table: %v", err)
	}
	if _, err := uuid.Parse(serverID); err != nil {
		return fmt.Errorf("servers table ID is not a valid UUID: %s", serverID)
	}

	// Check configs table
	var configID, configServerID string
	err = db.Raw("SELECT id, server_id FROM configs LIMIT 1").Row().Scan(&configID, &configServerID)
	if err != nil {
		return fmt.Errorf("failed to query configs table: %v", err)
	}
	if _, err := uuid.Parse(configID); err != nil {
		return fmt.Errorf("configs table ID is not a valid UUID: %s", configID)
	}
	if _, err := uuid.Parse(configServerID); err != nil {
		return fmt.Errorf("configs table server_id is not a valid UUID: %s", configServerID)
	}

	// Check state_histories table
	var stateID, stateServerID string
	err = db.Raw("SELECT id, server_id FROM state_histories LIMIT 1").Row().Scan(&stateID, &stateServerID)
	if err != nil {
		return fmt.Errorf("failed to query state_histories table: %v", err)
	}
	if _, err := uuid.Parse(stateID); err != nil {
		return fmt.Errorf("state_histories table ID is not a valid UUID: %s", stateID)
	}
	if _, err := uuid.Parse(stateServerID); err != nil {
		return fmt.Errorf("state_histories table server_id is not a valid UUID: %s", stateServerID)
	}

	// Check steam_credentials table
	var steamID string
	err = db.Raw("SELECT id FROM steam_credentials LIMIT 1").Scan(&steamID).Error
	if err != nil {
		return fmt.Errorf("failed to query steam_credentials table: %v", err)
	}
	if _, err := uuid.Parse(steamID); err != nil {
		return fmt.Errorf("steam_credentials table ID is not a valid UUID: %s", steamID)
	}

	// Check system_configs table
	var systemID string
	err = db.Raw("SELECT id FROM system_configs LIMIT 1").Scan(&systemID).Error
	if err != nil {
		return fmt.Errorf("failed to query system_configs table: %v", err)
	}
	if _, err := uuid.Parse(systemID); err != nil {
		return fmt.Errorf("system_configs table ID is not a valid UUID: %s", systemID)
	}

	logging.Info("✓ All tables successfully migrated to UUID primary keys")
	return nil
}

func testRoleSystem(db *gorm.DB) error {
	// Auto-migrate the models for role system
	db.AutoMigrate(&model.Role{}, &model.Permission{}, &model.User{})

	// Create repository and service
	repo := repository.NewMembershipRepository(db)
	service := service.NewMembershipService(repo)

	ctx := context.Background()

	// Setup initial data (this should create Super Admin, Admin, and Manager roles)
	if err := service.SetupInitialData(ctx); err != nil {
		return fmt.Errorf("failed to setup initial data: %v", err)
	}

	// Test that all three roles were created
	roles, err := service.GetAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get roles: %v", err)
	}

	expectedRoles := map[string]bool{
		"Super Admin": false,
		"Admin":       false,
		"Manager":     false,
	}

	for _, role := range roles {
		if _, exists := expectedRoles[role.Name]; exists {
			expectedRoles[role.Name] = true
		}
	}

	for roleName, found := range expectedRoles {
		if !found {
			return fmt.Errorf("role '%s' was not created", roleName)
		}
	}

	// Test permissions for each role
	superAdminRole, err := repo.FindRoleByName(ctx, "Super Admin")
	if err != nil {
		return fmt.Errorf("failed to find Super Admin role: %v", err)
	}

	adminRole, err := repo.FindRoleByName(ctx, "Admin")
	if err != nil {
		return fmt.Errorf("failed to find Admin role: %v", err)
	}

	managerRole, err := repo.FindRoleByName(ctx, "Manager")
	if err != nil {
		return fmt.Errorf("failed to find Manager role: %v", err)
	}

	// Load permissions for roles
	db.Preload("Permissions").Find(superAdminRole)
	db.Preload("Permissions").Find(adminRole)
	db.Preload("Permissions").Find(managerRole)

	// Super Admin and Admin should have all permissions
	allPermissions := model.AllPermissions()
	if len(superAdminRole.Permissions) != len(allPermissions) {
		return fmt.Errorf("Super Admin should have all %d permissions, but has %d", len(allPermissions), len(superAdminRole.Permissions))
	}

	if len(adminRole.Permissions) != len(allPermissions) {
		return fmt.Errorf("Admin should have all %d permissions, but has %d", len(allPermissions), len(adminRole.Permissions))
	}

	// Manager should have limited permissions (no create/delete for membership, role, user, server)
	expectedManagerPermissions := []string{
		model.ServerView,
		model.ServerUpdate,
		model.ServerStart,
		model.ServerStop,
		model.ConfigView,
		model.ConfigUpdate,
		model.UserView,
		model.RoleView,
		model.MembershipView,
	}

	if len(managerRole.Permissions) != len(expectedManagerPermissions) {
		return fmt.Errorf("Manager should have %d permissions, but has %d", len(expectedManagerPermissions), len(managerRole.Permissions))
	}

	// Verify Manager doesn't have restricted permissions
	restrictedPermissions := []string{
		model.ServerCreate,
		model.ServerDelete,
		model.UserCreate,
		model.UserDelete,
		model.RoleCreate,
		model.RoleDelete,
		model.MembershipCreate,
	}

	for _, restrictedPerm := range restrictedPermissions {
		for _, managerPerm := range managerRole.Permissions {
			if managerPerm.Name == restrictedPerm {
				return fmt.Errorf("Manager should not have permission '%s'", restrictedPerm)
			}
		}
	}

	logging.Info("✓ Role system working correctly")
	logging.Info("  - Super Admin role: %d permissions", len(superAdminRole.Permissions))
	logging.Info("  - Admin role: %d permissions", len(adminRole.Permissions))
	logging.Info("  - Manager role: %d permissions", len(managerRole.Permissions))

	return nil
}

func testSuperAdminDeletionPrevention(db *gorm.DB) error {
	// Create repository and service
	repo := repository.NewMembershipRepository(db)
	service := service.NewMembershipService(repo)

	ctx := context.Background()

	// Find the default admin user (should be Super Admin)
	adminUser, err := repo.FindUserByUsername(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to find admin user: %v", err)
	}

	// Try to delete the Super Admin user (should fail)
	err = service.DeleteUser(ctx, adminUser.ID)
	if err == nil {
		return fmt.Errorf("deleting Super Admin user should have failed, but it succeeded")
	}

	if err.Error() != "cannot delete Super Admin user" {
		return fmt.Errorf("expected 'cannot delete Super Admin user' error, got: %v", err)
	}

	logging.Info("✓ Super Admin deletion prevention working correctly")
	return nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
