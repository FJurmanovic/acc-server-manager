package controller

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// InitializeControllers
// Initializes Dependency Injection modules and registers controllers
//
//	Args:
//		*dig.Container: Dig Container
func InitializeControllers(c *dig.Container) {
	service.InitializeServices(c)

	err := c.Invoke(NewApiController)
	if err != nil {
		panic("unable to initialize api controller")
	}

	err = c.Invoke(NewConfigController)
	if err != nil {
		panic("unable to initialize config controller")
	}

	err = c.Invoke(NewServerController)
	if err != nil {
		panic("unable to initialize server controller")
	}

	err = c.Invoke(NewLookupController)
	if err != nil {
		panic("unable to initialize lookup controller")
	}
}

// FilteredResponse
// Gets query parameters and populates FilteredResponse model.
//
//	Args:
//		*gin.Context: Gin Application Context
//	Returns:
//		*model.FilteredResponse: Filtered response
func FilteredResponse(c *fiber.Ctx) *model.FilteredResponse {
	filtered := new(model.FilteredResponse)
	page := c.Params("page")
	rpp := c.Params("rpp")
	sortBy := c.Params("sortBy")

	dividers := [5]string{"|", " ", ".", "/", ","}

	for _, div := range dividers {
		sortArr := strings.Split(sortBy, div)

		if len(sortArr) >= 2 {
			sortBy = fmt.Sprintf("%s %s", common.ToSnakeCase(sortArr[0]), strings.ToUpper(sortArr[1]))
		}
	}

	filtered.Embed = c.Params("embed")
	filtered.Page, _ = strconv.Atoi(page)
	filtered.Rpp, _ = strconv.Atoi(rpp)
	filtered.SortBy = sortBy

	return filtered
}
