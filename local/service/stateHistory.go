package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"log"

	"github.com/gofiber/fiber/v2"
)

type StateHistoryService struct {
	repository *repository.StateHistoryRepository
}

func NewStateHistoryService(repository *repository.StateHistoryRepository) *StateHistoryService {
	return &StateHistoryService{
		repository: repository,
	}
}

// GetAll
// Gets All rows from StateHistory table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (s *StateHistoryService) GetAll(ctx *fiber.Ctx, filter *model.StateHistoryFilter) (*[]model.StateHistory, error) {
	result, err := s.repository.GetAll(ctx.UserContext(), filter)
	if err != nil {
		log.Printf("Error getting state history: %v", err)
		return nil, err
	}
	return result, nil
}

func (s *StateHistoryService) Insert(ctx *fiber.Ctx, model *model.StateHistory) error {
	if err := s.repository.Insert(ctx.UserContext(), model); err != nil {
		log.Printf("Error inserting state history: %v", err)
		return err
	}
	return nil
}