package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"

	"github.com/gofiber/fiber/v2"
)

type LookupService struct {
	repository *repository.LookupRepository
	cache     *model.LookupCache
}

func NewLookupService(repository *repository.LookupRepository) *LookupService {
	return &LookupService{
		repository: repository,
		cache:     model.NewLookupCache(),
	}
}

func (s *LookupService) GetTracks(ctx *fiber.Ctx) (interface{}, error) {
	if cached, exists := s.cache.Get("tracks"); exists {
		return cached, nil
	}

	tracks := s.repository.GetTracks(ctx.UserContext())
	s.cache.Set("tracks", tracks)
	return tracks, nil
}

func (s *LookupService) GetCarModels(ctx *fiber.Ctx) (interface{}, error) {
	if cached, exists := s.cache.Get("cars"); exists {
		return cached, nil
	}

	cars := s.repository.GetCarModels(ctx.UserContext())
	s.cache.Set("cars", cars)
	return cars, nil
}

func (s *LookupService) GetDriverCategories(ctx *fiber.Ctx) (interface{}, error) {
	if cached, exists := s.cache.Get("drivers"); exists {
		return cached, nil
	}

	categories := s.repository.GetDriverCategories(ctx.UserContext())
	s.cache.Set("drivers", categories)
	return categories, nil
}

func (s *LookupService) GetCupCategories(ctx *fiber.Ctx) (interface{}, error) {
	if cached, exists := s.cache.Get("cups"); exists {
		return cached, nil
	}

	categories := s.repository.GetCupCategories(ctx.UserContext())
	s.cache.Set("cups", categories)
	return categories, nil
}

func (s *LookupService) GetSessionTypes(ctx *fiber.Ctx) (interface{}, error) {
	if cached, exists := s.cache.Get("sessions"); exists {
		return cached, nil
	}

	types := s.repository.GetSessionTypes(ctx.UserContext())
	s.cache.Set("sessions", types)
	return types, nil
}

// ClearCache clears all cached lookup data
func (s *LookupService) ClearCache() {
	s.cache.Clear()
}

// PreloadCache loads all lookup data into cache
func (s *LookupService) PreloadCache(ctx *fiber.Ctx) {
	s.GetTracks(ctx)
	s.GetCarModels(ctx)
	s.GetDriverCategories(ctx)
	s.GetCupCategories(ctx)
	s.GetSessionTypes(ctx)
}
