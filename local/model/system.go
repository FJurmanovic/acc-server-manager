package model

type MigrateImageRequest struct {
	StopRunning bool `json:"stopRunning"`
}

type ServerMigrationEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	WasRunning bool   `json:"wasRunning"`
}

type ServerMigrationError struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Error string `json:"error"`
}

type MigrationResult struct {
	Image    string                 `json:"image"`
	Migrated []ServerMigrationEntry `json:"migrated"`
	Skipped  []ServerMigrationEntry `json:"skipped"`
	Failed   []ServerMigrationError `json:"failed"`
}
