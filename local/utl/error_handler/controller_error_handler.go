package error_handler

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ControllerErrorHandler provides centralized error handling for controllers
type ControllerErrorHandler struct {
	errorLogger *logging.ErrorLogger
}

// NewControllerErrorHandler creates a new controller error handler instance
func NewControllerErrorHandler() *ControllerErrorHandler {
	return &ControllerErrorHandler{
		errorLogger: logging.GetErrorLogger(),
	}
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// HandleError handles controller errors with logging and standardized responses
func (ceh *ControllerErrorHandler) HandleError(c *fiber.Ctx, err error, statusCode int, context ...string) error {
	if err == nil {
		return nil
	}

	// Get caller information for logging
	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, "acc-server-manager/")

	// Build context string
	contextStr := ""
	if len(context) > 0 {
		contextStr = fmt.Sprintf("[%s] ", strings.Join(context, "|"))
	}

	// Clean error message (remove null bytes)
	cleanErrorMsg := strings.ReplaceAll(err.Error(), "\x00", "")

	// Log the error with context
	ceh.errorLogger.LogWithContext(
		fmt.Sprintf("CONTROLLER_ERROR [%s:%d]", file, line),
		"%s%s",
		contextStr,
		cleanErrorMsg,
	)

	// Create standardized error response
	errorResponse := ErrorResponse{
		Error: cleanErrorMsg,
		Code:  statusCode,
	}

	// Add request details if available
	if c != nil {
		if errorResponse.Details == nil {
			errorResponse.Details = make(map[string]string)
		}
		
		// Safely extract request details
		func() {
			defer func() {
				if r := recover(); r != nil {
					// If any of these panic, just skip adding the details
					return
				}
			}()
			
			errorResponse.Details["method"] = c.Method()
			errorResponse.Details["path"] = c.Path()
			
			// Safely get IP address
			if ip := c.IP(); ip != "" {
				errorResponse.Details["ip"] = ip
			} else {
				errorResponse.Details["ip"] = "unknown"
			}
		}()
	}

	// Return appropriate response based on status code
	if c == nil {
		// If context is nil, we can't return a response
		return fmt.Errorf("cannot return HTTP response: context is nil")
	}
	
	if statusCode >= 500 {
		// For server errors, don't expose internal details
		return c.Status(statusCode).JSON(ErrorResponse{
			Error: "Internal server error",
			Code:  statusCode,
		})
	}

	return c.Status(statusCode).JSON(errorResponse)
}

// HandleValidationError handles validation errors specifically
func (ceh *ControllerErrorHandler) HandleValidationError(c *fiber.Ctx, err error, field string) error {
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "VALIDATION", field)
}

// HandleDatabaseError handles database-related errors
func (ceh *ControllerErrorHandler) HandleDatabaseError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusInternalServerError, "DATABASE")
}

// HandleAuthError handles authentication/authorization errors
func (ceh *ControllerErrorHandler) HandleAuthError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusUnauthorized, "AUTH")
}

// HandleNotFoundError handles resource not found errors
func (ceh *ControllerErrorHandler) HandleNotFoundError(c *fiber.Ctx, resource string) error {
	err := fmt.Errorf("%s not found", resource)
	return ceh.HandleError(c, err, fiber.StatusNotFound, "NOT_FOUND")
}

// HandleBusinessLogicError handles business logic errors
func (ceh *ControllerErrorHandler) HandleBusinessLogicError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "BUSINESS_LOGIC")
}

// HandleServiceError handles service layer errors
func (ceh *ControllerErrorHandler) HandleServiceError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusInternalServerError, "SERVICE")
}

// HandleParsingError handles request parsing errors
func (ceh *ControllerErrorHandler) HandleParsingError(c *fiber.Ctx, err error) error {
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "PARSING")
}

// HandleUUIDError handles UUID parsing errors
func (ceh *ControllerErrorHandler) HandleUUIDError(c *fiber.Ctx, field string) error {
	err := fmt.Errorf("invalid %s format", field)
	return ceh.HandleError(c, err, fiber.StatusBadRequest, "UUID_VALIDATION", field)
}

// Global controller error handler instance
var globalErrorHandler *ControllerErrorHandler

// GetControllerErrorHandler returns the global controller error handler instance
func GetControllerErrorHandler() *ControllerErrorHandler {
	if globalErrorHandler == nil {
		globalErrorHandler = NewControllerErrorHandler()
	}
	return globalErrorHandler
}

// Convenience functions using the global error handler

// HandleError handles controller errors using the global error handler
func HandleError(c *fiber.Ctx, err error, statusCode int, context ...string) error {
	return GetControllerErrorHandler().HandleError(c, err, statusCode, context...)
}

// HandleValidationError handles validation errors using the global error handler
func HandleValidationError(c *fiber.Ctx, err error, field string) error {
	return GetControllerErrorHandler().HandleValidationError(c, err, field)
}

// HandleDatabaseError handles database errors using the global error handler
func HandleDatabaseError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleDatabaseError(c, err)
}

// HandleAuthError handles auth errors using the global error handler
func HandleAuthError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleAuthError(c, err)
}

// HandleNotFoundError handles not found errors using the global error handler
func HandleNotFoundError(c *fiber.Ctx, resource string) error {
	return GetControllerErrorHandler().HandleNotFoundError(c, resource)
}

// HandleBusinessLogicError handles business logic errors using the global error handler
func HandleBusinessLogicError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleBusinessLogicError(c, err)
}

// HandleServiceError handles service errors using the global error handler
func HandleServiceError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleServiceError(c, err)
}

// HandleParsingError handles parsing errors using the global error handler
func HandleParsingError(c *fiber.Ctx, err error) error {
	return GetControllerErrorHandler().HandleParsingError(c, err)
}

// HandleUUIDError handles UUID errors using the global error handler
func HandleUUIDError(c *fiber.Ctx, field string) error {
	return GetControllerErrorHandler().HandleUUIDError(c, field)
}
