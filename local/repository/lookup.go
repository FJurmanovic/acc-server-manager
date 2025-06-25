package repository

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/cache"
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	cacheDuration            = 1 * time.Hour
	tracksCacheKey           = "tracks"
	carModelsCacheKey        = "carModels"
	driverCategoriesCacheKey = "driverCategories"
	cupCategoriesCacheKey    = "cupCategories"
	sessionTypesCacheKey     = "sessionTypes"
)

type LookupRepository struct {
	db    *gorm.DB
	cache *cache.InMemoryCache
}

func NewLookupRepository(db *gorm.DB, cache *cache.InMemoryCache) *LookupRepository {
	return &LookupRepository{
		db:    db,
		cache: cache,
	}
}

func (r *LookupRepository) GetTracks(ctx context.Context) (*[]model.Track, error) {
	fetcher := func() (*[]model.Track, error) {
		db := r.db.WithContext(ctx)
		items := new([]model.Track)
		if err := db.Find(items).Error; err != nil {
			return nil, err
		}
		return items, nil
	}
	return cache.GetOrSet(r.cache, tracksCacheKey, cacheDuration, fetcher)
}

func (r *LookupRepository) GetCarModels(ctx context.Context) (*[]model.CarModel, error) {
	fetcher := func() (*[]model.CarModel, error) {
		db := r.db.WithContext(ctx)
		items := new([]model.CarModel)
		if err := db.Find(items).Error; err != nil {
			return nil, err
		}
		return items, nil
	}
	return cache.GetOrSet(r.cache, carModelsCacheKey, cacheDuration, fetcher)
}

func (r *LookupRepository) GetDriverCategories(ctx context.Context) (*[]model.DriverCategory, error) {
	fetcher := func() (*[]model.DriverCategory, error) {
		db := r.db.WithContext(ctx)
		items := new([]model.DriverCategory)
		if err := db.Find(items).Error; err != nil {
			return nil, err
		}
		return items, nil
	}
	return cache.GetOrSet(r.cache, driverCategoriesCacheKey, cacheDuration, fetcher)
}

func (r *LookupRepository) GetCupCategories(ctx context.Context) (*[]model.CupCategory, error) {
	fetcher := func() (*[]model.CupCategory, error) {
		db := r.db.WithContext(ctx)
		items := new([]model.CupCategory)
		if err := db.Find(items).Error; err != nil {
			return nil, err
		}
		return items, nil
	}
	return cache.GetOrSet(r.cache, cupCategoriesCacheKey, cacheDuration, fetcher)
}

func (r *LookupRepository) GetSessionTypes(ctx context.Context) (*[]model.SessionType, error) {
	fetcher := func() (*[]model.SessionType, error) {
		db := r.db.WithContext(ctx)
		items := new([]model.SessionType)
		if err := db.Find(items).Error; err != nil {
			return nil, err
		}
		return items, nil
	}
	return cache.GetOrSet(r.cache, sessionTypesCacheKey, cacheDuration, fetcher)
}
