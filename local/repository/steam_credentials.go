package repository

import (
	"acc-server-manager/local/model"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SteamCredentialsRepository struct {
	db *gorm.DB
}

func NewSteamCredentialsRepository(db *gorm.DB) *SteamCredentialsRepository {
	return &SteamCredentialsRepository{
		db: db,
	}
}

func (r *SteamCredentialsRepository) GetCurrent(ctx context.Context) (*model.SteamCredentials, error) {
	var creds model.SteamCredentials
	result := r.db.WithContext(ctx).Order("id desc").First(&creds)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &creds, nil
}

func (r *SteamCredentialsRepository) Save(ctx context.Context, creds *model.SteamCredentials) error {
	if creds.ID == uuid.Nil {
		return r.db.WithContext(ctx).Create(creds).Error
	}
	return r.db.WithContext(ctx).Save(creds).Error
}

func (r *SteamCredentialsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.SteamCredentials{}, id).Error
}
