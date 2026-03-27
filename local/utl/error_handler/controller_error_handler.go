package error_handler

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ControllerErrorHandler struct {
	errorLogger *logging.ErrorLogger
}

func NewControllerErrorHandler() *ControllerErrorHandler {
	return &ControllerErrorHandler{
		errorLogger: logging.GetErrorLogger(),
	}
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

func (ceh *ControllerErrorHandler) HandleError(c *fiber.Ctx, err error, statusCode int, context ...string) error {
	if err == nil {
		return nil
	}

	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, "acc-server-manager/")

	contextStr := ""
	if len(context) > 0 {
		contextStr = fmt.Sprintf("[%s] ", strings.Join(context, "|"))
	}

	cleanErrorMsg := strings.ReplaceAll(err.Error(), "\x00", "")

	ceh.errorLogger.LogWithContext(
		fmt.Sprintf("CONTROLLER_ERROR [%s:%d]", file, line),
		"%s%s",
		contextStr,
		cleanErrorMsg,
	)

	errorResponse := ErrorResponse{
		Error: cleanErrorMsg,
		Code:  statusCode,
	}

	if c != nil {
		if errorResponse.Details == nil {
			errorResponse.Details = make(map[string]string)
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					return
				}
			}()

			errorResponse.Details["method"] = c.Method()
			errorResponse.Details["path"] = c.Path()

			if ip := c.IP(); ip != "" {
				errorResponse.Details["ip"] = ip
			} else {
				errorResponse.Details["ip"] = "unknown"
			}
		}()
	}

	if c == nil {
		return fmt.Errorf("cannot return HTTP response: context is nil")
	}

	if statusCode >= 500 {
		return c.Status(statusCode).JSON(ErrorResponse{
			Error: "Internal server error",
			Code:  statusCode,
		})
	}

	return c.Status(statusCode).JSON(errorResponse)
}

func (ceh *ControllerErrorHandler) HandleValidationError(c *fiber.Ctx, err error, field string) error {
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "VALIDATION", field)
}

func (ceh *ControllerErrorHandler) HandleDatabaseError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusInternalServerError, "DATABASE")
}

func (ceh *ControllerErrorHandler) HandleAuthError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusUnauthorized, "AUTH")
}

func (ceh *ControllerErrorHandler) HandleNotFoundError(c *fiber.Ctx, resource string) error {
	err := fmt.Errorf("%s not found", resource)
	return ceh.HandleError(c, err, fiber.StatusNotFound, "NOT_FOUND")
}

func (ceh *ControllerErrorHandler) HandleBusinessLogicError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "BUSINESS_LOGIC")
}

func (ceh *ControllerErrorHandler) HandleServiceError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusInternalServerError, "SERVICE")
}

func (ceh *ControllerErrorHandler) HandleParsingError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "PARSING")
}

func (ceh *ControllerErrorHandler) HandleUUIDError(c *fiber.Ctx, field string) error {
	err := fmt.Errorf("invalid %s format", field)
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "UUID_VALIDATION", field)
}

var globalErrorHandler *ControllerErrorHandler

func GetControllerErrorHandler() *ControllerErrorHandler {
	if globalErrorHandler == nil {
		globalErrorHandler = NewControllerErrorHandler()
	}
	return globalErrorHandler
}

func HandleError(c *fiber.Ctx, err error, statusCode int, context ...string) error {
	return GetControllerErrorHandler().HandleError(c, err, statusCode, context...)
}

func HandleValidationError(c *fiber.Ctx, err error, field string) error {
	return GetControllerErrorHandler().HandleValidationError(c, err, field)
}

func HandleDatabaseError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleDatabaseError(c, err)
}

func HandleAuthError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleAuthError(c, err)
}

func HandleNotFoundError(c *fiber.Ctx, resource string) error {
	return GetControllerErrorHandler().HandleNotFoundError(c, resource)
}

func HandleBusinessLogicError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleBusinessLogicError(c, err)
}

func HandleServiceError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleServiceError(c, err)
}

func HandleParsingError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleParsingError(c, err)
}

func HandleUUIDError(c *fiber.Ctx, field string) error {
	return GetControllerErrorHandler().HandleUUIDError(c, field)
}
