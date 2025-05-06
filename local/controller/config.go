package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"log"

	"github.com/gofiber/fiber/v2"
)

type ConfigController struct {
	service    *service.ConfigService
	apiService *service.ApiService
}

// NewConfigController
// Initializes ConfigController.
//
//	Args:
//		*services.ConfigService: Config service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*ConfigController: Controller for "Config" interactions
func NewConfigController(as *service.ConfigService, routeGroups *common.RouteGroups, as2 *service.ApiService) *ConfigController {
	ac := &ConfigController{
		service:    as,
		apiService: as2,
	}

	routeGroups.Config.Put("/:file", ac.updateConfig)
	routeGroups.Config.Get("/:file", ac.getConfig)
	routeGroups.Config.Get("/", ac.getConfigs)

	return ac
}

// updateConfig returns Config
//
//	@Summary		Update config
//	@Description	Updates config
//	@Param			id path number true "required"
//	@Param			file path string true "required"
//	@Param			content body string true "required"
//	@Tags			Config
//	@Success		200	{array}	string
//	@Router			/v1/server/{id}/config/{file} [put]
func (ac *ConfigController) updateConfig(c *fiber.Ctx) error {
	restart := c.QueryBool("restart")
	serverID, _ := c.ParamsInt("id")
	c.Locals("serverId", serverID)

	var config map[string]interface{}
	if err := c.BodyParser(&config); err != nil {
		log.Print("Invalid config format")
		return c.Status(400).JSON(fiber.Map{"error": "Invalid config format"})
	}

	ConfigModel, err := ac.service.UpdateConfig(c, &config)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}
	if restart {
		ac.apiService.ApiRestartServer(c)
	}

	return c.JSON(ConfigModel)
}

// getConfig returns Config
//
//	@Summary		Return Config file
//	@Description	Returns Config file
//	@Param			id path number true "required"
//	@Param			file path string true "required"
//	@Tags			Config
//	@Success		200	{array}	string
//	@Router			/v1/server/{id}/config/{file} [get]
func (ac *ConfigController) getConfig(c *fiber.Ctx) error {
	Model, err := ac.service.GetConfig(c)
	if err != nil {
		log.Print(err.Error())
		return c.Status(400).SendString(err.Error())
	}
	return c.JSON(Model)
}

// getConfigs returns Config
//
//	@Summary		Return Configs
//	@Description	Return Config files
//	@Param			id path number true "required"
//	@Tags			Config
//	@Success		200	{array}	string
//	@Router			/v1/server/{id}/config [get]
func (ac *ConfigController) getConfigs(c *fiber.Ctx) error {
	Model, err := ac.service.GetConfigs(c)
	if err != nil {
		log.Print(err.Error())
		return c.Status(400).SendString(err.Error())
	}
	return c.JSON(Model)
}
