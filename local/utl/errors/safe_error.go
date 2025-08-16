package errors

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"os"
)

type SafeError struct {
	Message string
	Code    int
	Fatal   bool
}

func (e *SafeError) Error() string {
	return e.Message
}

func NewSafeError(message string, code int) *SafeError {
	return &SafeError{
		Message: message,
		Code:    code,
		Fatal:   false,
	}
}

func NewFatalError(message string, code int) *SafeError {
	return &SafeError{
		Message: message,
		Code:    code,
		Fatal:   true,
	}
}

func HandleError(err error, context string) {
	if err == nil {
		return
	}

	if safeErr, ok := err.(*SafeError); ok {
		if safeErr.Fatal {
			logging.Error("Fatal error in %s: %s", context, safeErr.Message)
			if os.Getenv("ENVIRONMENT") == "production" {
				logging.Error("Application shutting down due to fatal error")
				os.Exit(safeErr.Code)
			} else {
				logging.Warn("Fatal error occurred but not exiting in non-production environment")
			}
		} else {
			logging.Error("Error in %s: %s", context, safeErr.Message)
		}
	} else {
		logging.Error("Unexpected error in %s: %v", context, err)
	}
}

func SafeFatal(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	err := NewFatalError(formattedMessage, 1)
	HandleError(err, "application")
}

func SafeLog(message string, args ...interface{}) {
	formattedMessage := fmt.Sprintf(message, args...)
	err := NewSafeError(formattedMessage, 0)
	HandleError(err, "application")
}