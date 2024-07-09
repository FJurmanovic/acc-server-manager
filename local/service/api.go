package service

import (
	"acc-server-manager/local/utl/configs"

	"github.com/gofiber/fiber/v2"
)

type ApiService struct {
}

func NewApiService() *ApiService {
	return &ApiService{}
}

/*
GetFirst

Gets first row from API table.

	   	Args:
	   		context.Context: Application context
		Returns:
			string: Application version
*/
func (as ApiService) GetFirst(ctx *fiber.Ctx) string {
	return configs.Version
}
