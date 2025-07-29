package logging

import (
	"fmt"
	"sync"
)

// InfoLogger handles info-level logging
type InfoLogger struct {
	base *BaseLogger
}

// NewInfoLogger creates a new info logger instance
func NewInfoLogger() *InfoLogger {
	base, _ := InitializeBase("info")
	return &InfoLogger{
		base: base,
	}
}

// Log writes an info-level log entry
func (il *InfoLogger) Log(format string, v ...interface{}) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, format, v...)
	}
}

// LogWithContext writes an info-level log entry with additional context
func (il *InfoLogger) LogWithContext(context string, format string, v ...interface{}) {
	if il.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		il.base.Log(LogLevelInfo, contextualFormat, v...)
	}
}

// LogStartup logs application startup information
func (il *InfoLogger) LogStartup(component string, message string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "STARTUP [%s]: %s", component, message)
	}
}

// LogShutdown logs application shutdown information
func (il *InfoLogger) LogShutdown(component string, message string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "SHUTDOWN [%s]: %s", component, message)
	}
}

// LogOperation logs general operation information
func (il *InfoLogger) LogOperation(operation string, details string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "OPERATION [%s]: %s", operation, details)
	}
}

// LogStatus logs status changes or updates
func (il *InfoLogger) LogStatus(component string, status string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "STATUS [%s]: %s", component, status)
	}
}

// LogRequest logs incoming requests
func (il *InfoLogger) LogRequest(method string, path string, userAgent string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "REQUEST [%s %s] User-Agent: %s", method, path, userAgent)
	}
}

// LogResponse logs outgoing responses
func (il *InfoLogger) LogResponse(method string, path string, statusCode int, duration string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "RESPONSE [%s %s] Status: %d, Duration: %s", method, path, statusCode, duration)
	}
}

// Global info logger instance
var (
	infoLogger *InfoLogger
	infoOnce   sync.Once
)

// GetInfoLogger returns the global info logger instance
func GetInfoLogger() *InfoLogger {
	infoOnce.Do(func() {
		infoLogger = NewInfoLogger()
	})
	return infoLogger
}

// Info logs an info-level message using the global info logger
func Info(format string, v ...interface{}) {
	GetInfoLogger().Log(format, v...)
}

// InfoWithContext logs an info-level message with context using the global info logger
func InfoWithContext(context string, format string, v ...interface{}) {
	GetInfoLogger().LogWithContext(context, format, v...)
}

// InfoStartup logs application startup information using the global info logger
func InfoStartup(component string, message string) {
	GetInfoLogger().LogStartup(component, message)
}

// InfoShutdown logs application shutdown information using the global info logger
func InfoShutdown(component string, message string) {
	GetInfoLogger().LogShutdown(component, message)
}

// InfoOperation logs general operation information using the global info logger
func InfoOperation(operation string, details string) {
	GetInfoLogger().LogOperation(operation, details)
}

// InfoStatus logs status changes or updates using the global info logger
func InfoStatus(component string, status string) {
	GetInfoLogger().LogStatus(component, status)
}

// InfoRequest logs incoming requests using the global info logger
func InfoRequest(method string, path string, userAgent string) {
	GetInfoLogger().LogRequest(method, path, userAgent)
}

// InfoResponse logs outgoing responses using the global info logger
func InfoResponse(method string, path string, statusCode int, duration string) {
	GetInfoLogger().LogResponse(method, path, statusCode, duration)
}
