package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"

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
func (as StateHistoryService) GetAll(ctx *fiber.Ctx, id int) *[]model.StateHistory {
	return as.repository.GetAll(ctx.UserContext(), id)
}