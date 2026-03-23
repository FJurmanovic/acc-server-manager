package logging

import (
	"fmt"
	"sync"
)

type InfoLogger struct {
	base *BaseLogger
}

func NewInfoLogger() *InfoLogger {
	base, _ := InitializeBase("info")
	return &InfoLogger{
		base: base,
	}
}

func (il *InfoLogger) Log(format string, v ...interface{}) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, format, v...)
	}
}

func (il *InfoLogger) LogWithContext(context string, format string, v ...interface{}) {
	if il.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		il.base.Log(LogLevelInfo, contextualFormat, v...)
	}
}

func (il *InfoLogger) LogStartup(component string, message string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "STARTUP [%s]: %s", component, message)
	}
}

func (il *InfoLogger) LogShutdown(component string, message string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "SHUTDOWN [%s]: %s", component, message)
	}
}

func (il *InfoLogger) LogOperation(operation string, details string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "OPERATION [%s]: %s", operation, details)
	}
}

func (il *InfoLogger) LogStatus(component string, status string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "STATUS [%s]: %s", component, status)
	}
}

func (il *InfoLogger) LogRequest(method string, path string, userAgent string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "REQUEST [%s %s] User-Agent: %s", method, path, userAgent)
	}
}

func (il *InfoLogger) LogResponse(method string, path string, statusCode int, duration string) {
	if il.base != nil {
		il.base.Log(LogLevelInfo, "RESPONSE [%s %s] Status: %d, Duration: %s", method, path, statusCode, duration)
	}
}

var (
	infoLogger *InfoLogger
	infoOnce   sync.Once
)

func GetInfoLogger() *InfoLogger {
	infoOnce.Do(func() {
		infoLogger = NewInfoLogger()
	})
	return infoLogger
}

func Info(format string, v ...interface{}) {
	GetInfoLogger().Log(format, v...)
}

func InfoWithContext(context string, format string, v ...interface{}) {
	GetInfoLogger().LogWithContext(context, format, v...)
}

func InfoStartup(component string, message string) {
	GetInfoLogger().LogStartup(component, message)
}

func InfoShutdown(component string, message string) {
	GetInfoLogger().LogShutdown(component, message)
}

func InfoOperation(operation string, details string) {
	GetInfoLogger().LogOperation(operation, details)
}

func InfoStatus(component string, status string) {
	GetInfoLogger().LogStatus(component, status)
}

func InfoRequest(method string, path string, userAgent string) {
	GetInfoLogger().LogRequest(method, path, userAgent)
}

func InfoResponse(method string, path string, statusCode int, duration string) {
	GetInfoLogger().LogResponse(method, path, statusCode, duration)
}
