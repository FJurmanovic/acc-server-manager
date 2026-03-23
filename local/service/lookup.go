package service

import (
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"

	"github.com/gofiber/fiber/v2"
)

type LookupService struct {
	repository *repository.LookupRepository
}

func NewLookupService(repository *repository.LookupRepository) *LookupService {
	logging.Debug("Initializing LookupService")
	return &LookupService{
		repository: repository,
	}
}

func (s *LookupService) GetTracks(ctx *fiber.Ctx) (interface{}, error) {
	logging.Debug("Getting tracks")
	return s.repository.GetTracks(ctx.UserContext())
}

func (s *LookupService) GetCarModels(ctx *fiber.Ctx) (interface{}, error) {
	logging.Debug("Getting car models")
	return s.repository.GetCarModels(ctx.UserContext())
}

func (s *LookupService) GetDriverCategories(ctx *fiber.Ctx) (interface{}, error) {
	logging.Debug("Getting driver categories")
	return s.repository.GetDriverCategories(ctx.UserContext())
}

func (s *LookupService) GetCupCategories(ctx *fiber.Ctx) (interface{}, error) {
	logging.Debug("Getting cup categories")
	return s.repository.GetCupCategories(ctx.UserContext())
}

func (s *LookupService) GetSessionTypes(ctx *fiber.Ctx) (interface{}, error) {
	logging.Debug("Getting session types")
	return s.repository.GetSessionTypes(ctx.UserContext())
}
