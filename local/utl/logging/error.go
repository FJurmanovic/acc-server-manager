package logging

import (
	"fmt"
	"runtime"
)

// ErrorLogger handles error-level logging
type ErrorLogger struct {
	base *BaseLogger
}

// NewErrorLogger creates a new error logger instance
func NewErrorLogger() *ErrorLogger {
	return &ErrorLogger{
		base: GetBaseLogger("error"),
	}
}

// Log writes an error-level log entry
func (el *ErrorLogger) Log(format string, v ...interface{}) {
	if el.base != nil {
		el.base.Log(LogLevelError, format, v...)
	}
}

// LogWithContext writes an error-level log entry with additional context
func (el *ErrorLogger) LogWithContext(context string, format string, v ...interface{}) {
	if el.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		el.base.Log(LogLevelError, contextualFormat, v...)
	}
}

// LogError logs an error object with optional message
func (el *ErrorLogger) LogError(err error, message ...string) {
	if el.base != nil && err != nil {
		if len(message) > 0 {
			el.base.Log(LogLevelError, "%s: %v", message[0], err)
		} else {
			el.base.Log(LogLevelError, "Error: %v", err)
		}
	}
}

// LogWithStackTrace logs an error with stack trace
func (el *ErrorLogger) LogWithStackTrace(format string, v ...interface{}) {
	if el.base != nil {
		// Get stack trace
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		stackTrace := string(buf[:n])

		msg := fmt.Sprintf(format, v...)
		el.base.Log(LogLevelError, "%s\nStack Trace:\n%s", msg, stackTrace)
	}
}

// LogFatal logs a fatal error and exits the program
func (el *ErrorLogger) LogFatal(format string, v ...interface{}) {
	if el.base != nil {
		el.base.Log(LogLevelError, "[FATAL] "+format, v...)
		panic(fmt.Sprintf(format, v...))
	}
}

// Global error logger instance
var errorLogger *ErrorLogger

// GetErrorLogger returns the global error logger instance
func GetErrorLogger() *ErrorLogger {
	if errorLogger == nil {
		errorLogger = NewErrorLogger()
	}
	return errorLogger
}

// Error logs an error-level message using the global error logger
func Error(format string, v ...interface{}) {
	GetErrorLogger().Log(format, v...)
}

// ErrorWithContext logs an error-level message with context using the global error logger
func ErrorWithContext(context string, format string, v ...interface{}) {
	GetErrorLogger().LogWithContext(context, format, v...)
}

// LogError logs an error object using the global error logger
func LogError(err error, message ...string) {
	GetErrorLogger().LogError(err, message...)
}

// ErrorWithStackTrace logs an error with stack trace using the global error logger
func ErrorWithStackTrace(format string, v ...interface{}) {
	GetErrorLogger().LogWithStackTrace(format, v...)
}

// Fatal logs a fatal error and exits the program using the global error logger
func Fatal(format string, v ...interface{}) {
	GetErrorLogger().LogFatal(format, v...)
}
