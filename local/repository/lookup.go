package repository

import (
	"acc-server-manager/local/model"
	"context"

	"gorm.io/gorm"
)

type LookupRepository struct {
	db *gorm.DB
}

func NewLookupRepository(db *gorm.DB) *LookupRepository {
	return &LookupRepository{
		db: db,
	}
}

// GetTracks
// Gets Tracks rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupRepository) GetTracks(ctx context.Context) *[]model.Track {
	db := as.db.WithContext(ctx)
	TrackModel := new([]model.Track)
	db.Find(&TrackModel)
	return TrackModel
}

// GetCarModels
// Gets CarModels rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupRepository) GetCarModels(ctx context.Context) *[]model.CarModel {
	db := as.db.WithContext(ctx)
	CarModelModel := new([]model.CarModel)
	db.Find(&CarModelModel)
	return CarModelModel
}

// GetDriverCategories
// Gets DriverCategories rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupRepository) GetDriverCategories(ctx context.Context) *[]model.DriverCategory {
	db := as.db.WithContext(ctx)
	DriverCategoryModel := new([]model.DriverCategory)
	db.Find(&DriverCategoryModel)
	return DriverCategoryModel
}

// GetCupCategories
// Gets CupCategories rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupRepository) GetCupCategories(ctx context.Context) *[]model.CupCategory {
	db := as.db.WithContext(ctx)
	CupCategoryModel := new([]model.CupCategory)
	db.Find(&CupCategoryModel)
	return CupCategoryModel
}

// GetSessionTypes
// Gets SessionTypes rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupRepository) GetSessionTypes(ctx context.Context) *[]model.SessionType {
	db := as.db.WithContext(ctx)
	SessionTypesModel := new([]model.SessionType)
	db.Find(&SessionTypesModel)
	return SessionTypesModel
}
