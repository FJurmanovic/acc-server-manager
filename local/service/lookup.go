package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"

	"github.com/gofiber/fiber/v2"
)

type LookupService struct {
	repository *repository.LookupRepository
}

func NewLookupService(repository *repository.LookupRepository) *LookupService {
	return &LookupService{
		repository: repository,
	}
}

// GetTracks
// Gets Tracks rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as LookupService) GetTracks(ctx *fiber.Ctx) *[]model.Track {
	return as.repository.GetTracks(ctx.UserContext())
}

// GetCarModels
// Gets CarModels rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupService) GetCarModels(ctx *fiber.Ctx) *[]model.CarModel {
	return as.repository.GetCarModels(ctx.UserContext())
}

// GetDriverCategories
// Gets DriverCategories rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupService) GetDriverCategories(ctx *fiber.Ctx) *[]model.DriverCategory {
	return as.repository.GetDriverCategories(ctx.UserContext())
}

// GetCupCategories
// Gets CupCategories rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupService) GetCupCategories(ctx *fiber.Ctx) *[]model.CupCategory {
	return as.repository.GetCupCategories(ctx.UserContext())
}

// GetSessionTypes
// Gets SessionTypes rows from Lookup table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.LookupModel: Lookup object from database.
func (as LookupService) GetSessionTypes(ctx *fiber.Ctx) *[]model.SessionType {
	return as.repository.GetSessionTypes(ctx.UserContext())
}
