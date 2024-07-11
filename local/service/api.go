package service

import (
	"acc-server-manager/local/utl/configs"
	"os/exec"

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

func (as ApiService) StartServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	print(service)
	if !ok {
		return "", fiber.NewError(400)
	}
	cmd := exec.Command("sc", "start", service)
	output, err := cmd.CombinedOutput()
	print(string(output[:]))
	if err != nil {
		return "", fiber.NewError(500)
	}
	return string(output[:]), nil

}
