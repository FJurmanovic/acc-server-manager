package repository

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ServerRepository struct {
	*BaseRepository[model.Server, model.ServerFilter]
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	repo := &ServerRepository{
		BaseRepository: NewBaseRepository[model.Server, model.ServerFilter](db, model.Server{}),
	}

	// Run migrations
	if err := repo.migrateServerTable(); err != nil {
		panic(err)
	}

	return repo
}

// migrateServerTable ensures all required columns exist with proper defaults
func (r *ServerRepository) migrateServerTable() error {
	// Create a temporary table with all required columns
	if err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS servers_new (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			ip TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 9600,
			path TEXT NOT NULL,
			service_name TEXT NOT NULL,
			date_created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			from_steam_cmd BOOLEAN NOT NULL DEFAULT 1
		)
	`).Error; err != nil {
		return err
	}

	// Copy data from old table, setting defaults for new columns
	if err := r.db.Exec(`
		INSERT INTO servers_new (
			id, 
			name, 
			ip, 
			port,
			path, 
			service_name,
			date_created,
			from_steam_cmd
		)
		SELECT 
			id,
			COALESCE(name, 'Server ' || id) as name,
			COALESCE(ip, '127.0.0.1') as ip,
			COALESCE(port, 9600) as port,
			path,
			COALESCE(service_name, 'ACC-Server-' || id) as service_name,
			COALESCE(date_created, CURRENT_TIMESTAMP) as date_created,
			COALESCE(from_steam_cmd, 1) as from_steam_cmd
		FROM servers
	`).Error; err != nil {
		// If the old table doesn't exist, this is a fresh install
		if err := r.db.Exec(`DROP TABLE IF EXISTS servers_new`).Error; err != nil {
			return err
		}
		return nil
	}

	// Replace old table with new one
	if err := r.db.Exec(`DROP TABLE IF EXISTS servers`).Error; err != nil {
		return err
	}
	if err := r.db.Exec(`ALTER TABLE servers_new RENAME TO servers`).Error; err != nil {
		return err
	}

	return nil
}

// GetFirstByServiceName
// Gets first row from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ServerModel: Server object from database.
func (r *ServerRepository) GetFirstByServiceName(ctx context.Context, serviceName string) (*model.Server, error) {
	result := new(model.Server)
	if err := r.db.WithContext(ctx).Where("service_name = ?", serviceName).First(result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return result, nil
}