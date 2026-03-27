package logging

import (
	"fmt"
	"sync"
)

var (
	logger *Logger
	once   sync.Once
)

type Logger struct {
	base        *BaseLogger
	errorLogger *ErrorLogger
	warnLogger  *WarnLogger
	infoLogger  *InfoLogger
	debugLogger *DebugLogger
}

func Initialize() (*Logger, error) {
	var err error
	once.Do(func() {
		logger, err = newLogger()
	})
	return logger, err
}

func newLogger() (*Logger, error) {
	baseLogger, err := InitializeBase("log")
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		base:        baseLogger,
		errorLogger: GetErrorLogger(),
		warnLogger:  GetWarnLogger(),
		infoLogger:  GetInfoLogger(),
		debugLogger: GetDebugLogger(),
	}

	return logger, nil
}

func (l *Logger) Close() error {
	if l.base != nil {
		return l.base.Close()
	}
	return nil
}

func (l *Logger) log(level, format string, v ...interface{}) {
	if l.base != nil {
		l.base.LogWithCaller(LogLevel(level), 3, format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.infoLogger != nil {
		l.infoLogger.Log(format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.errorLogger != nil {
		l.errorLogger.Log(format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.warnLogger != nil {
		l.warnLogger.Log(format, v...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.debugLogger != nil {
		l.debugLogger.Log(format, v...)
	}
}

func (l *Logger) Panic(format string) {
	if l.errorLogger != nil {
		l.errorLogger.LogFatal(format)
	}
}

func LegacyInfo(format string, v ...interface{}) {
	if logger != nil {
		logger.Info(format, v...)
	} else {
		GetInfoLogger().Log(format, v...)
	}
}

func LegacyError(format string, v ...interface{}) {
	if logger != nil {
		logger.Error(format, v...)
	} else {
		GetErrorLogger().Log(format, v...)
	}
}

func LegacyWarn(format string, v ...interface{}) {
	if logger != nil {
		logger.Warn(format, v...)
	} else {
		GetWarnLogger().Log(format, v...)
	}
}

func LegacyDebug(format string, v ...interface{}) {
	if logger != nil {
		logger.Debug(format, v...)
	} else {
		GetDebugLogger().Log(format, v...)
	}
}

func Panic(format string) {
	if logger != nil {
		logger.Panic(format)
	} else {
		GetErrorLogger().LogFatal(format)
	}
}

func LogStartup(component string, message string) {
	GetInfoLogger().LogStartup(component, message)
}

func LogShutdown(component string, message string) {
	GetInfoLogger().LogShutdown(component, message)
}

func LogOperation(operation string, details string) {
	GetInfoLogger().LogOperation(operation, details)
}

func LogRequest(method string, path string, userAgent string) {
	GetInfoLogger().LogRequest(method, path, userAgent)
}

func LogResponse(method string, path string, statusCode int, duration string) {
	GetInfoLogger().LogResponse(method, path, statusCode, duration)
}

func LogSQL(query string, args ...interface{}) {
	GetDebugLogger().LogSQL(query, args...)
}

func LogMemory() {
	GetDebugLogger().LogMemory()
}

func LogTiming(operation string, duration interface{}) {
	GetDebugLogger().LogTiming(operation, duration)
}

func GetLegacyLogger() *Logger {
	if logger == nil {
		logger, _ = Initialize()
	}
	return logger
}

func InitializeLogging() error {
	_, err := Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize legacy logger: %v", err)
	}

	GetErrorLogger()
	GetWarnLogger()
	GetInfoLogger()
	GetDebugLogger()

	Info("Logging system initialized successfully")

	return nil
}
