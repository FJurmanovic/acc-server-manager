package logging

import (
	"fmt"
	"sync"
)

var (
	// Legacy logger for backward compatibility
	logger *Logger
	once   sync.Once
)

// Logger maintains backward compatibility with existing code
type Logger struct {
	base        *BaseLogger
	errorLogger *ErrorLogger
	warnLogger  *WarnLogger
	infoLogger  *InfoLogger
	debugLogger *DebugLogger
}

// Initialize creates or gets the singleton logger instance
// This maintains backward compatibility with existing code
func Initialize() (*Logger, error) {
	var err error
	once.Do(func() {
		logger, err = newLogger()
	})
	return logger, err
}

func newLogger() (*Logger, error) {
	// Initialize the base logger
	baseLogger, err := InitializeBase("log")
	if err != nil {
		return nil, err
	}

	// Create the legacy logger wrapper
	logger := &Logger{
		base:        baseLogger,
		errorLogger: GetErrorLogger(),
		warnLogger:  GetWarnLogger(),
		infoLogger:  GetInfoLogger(),
		debugLogger: GetDebugLogger(),
	}

	return logger, nil
}

// Close closes the logger
func (l *Logger) Close() error {
	if l.base != nil {
		return l.base.Close()
	}
	return nil
}

// Legacy methods for backward compatibility
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

// Global convenience functions for backward compatibility
// These are now implemented in individual logger files to avoid redeclaration
func LegacyInfo(format string, v ...interface{}) {
	if logger != nil {
		logger.Info(format, v...)
	} else {
		// Fallback to direct logger if legacy logger not initialized
		GetInfoLogger().Log(format, v...)
	}
}

func LegacyError(format string, v ...interface{}) {
	if logger != nil {
		logger.Error(format, v...)
	} else {
		// Fallback to direct logger if legacy logger not initialized
		GetErrorLogger().Log(format, v...)
	}
}

func LegacyWarn(format string, v ...interface{}) {
	if logger != nil {
		logger.Warn(format, v...)
	} else {
		// Fallback to direct logger if legacy logger not initialized
		GetWarnLogger().Log(format, v...)
	}
}

func LegacyDebug(format string, v ...interface{}) {
	if logger != nil {
		logger.Debug(format, v...)
	} else {
		// Fallback to direct logger if legacy logger not initialized
		GetDebugLogger().Log(format, v...)
	}
}

func Panic(format string) {
	if logger != nil {
		logger.Panic(format)
	} else {
		// Fallback to direct logger if legacy logger not initialized
		GetErrorLogger().LogFatal(format)
	}
}

// Enhanced logging convenience functions
// These provide direct access to specialized logging functions

// LogStartup logs application startup information
func LogStartup(component string, message string) {
	GetInfoLogger().LogStartup(component, message)
}

// LogShutdown logs application shutdown information
func LogShutdown(component string, message string) {
	GetInfoLogger().LogShutdown(component, message)
}

// LogOperation logs general operation information
func LogOperation(operation string, details string) {
	GetInfoLogger().LogOperation(operation, details)
}

// LogRequest logs incoming HTTP requests
func LogRequest(method string, path string, userAgent string) {
	GetInfoLogger().LogRequest(method, path, userAgent)
}

// LogResponse logs outgoing HTTP responses
func LogResponse(method string, path string, statusCode int, duration string) {
	GetInfoLogger().LogResponse(method, path, statusCode, duration)
}

// LogSQL logs SQL queries for debugging
func LogSQL(query string, args ...interface{}) {
	GetDebugLogger().LogSQL(query, args...)
}

// LogMemory logs memory usage information
func LogMemory() {
	GetDebugLogger().LogMemory()
}

// LogTiming logs timing information for performance debugging
func LogTiming(operation string, duration interface{}) {
	GetDebugLogger().LogTiming(operation, duration)
}

// GetLegacyLogger returns the legacy logger instance for backward compatibility
func GetLegacyLogger() *Logger {
	if logger == nil {
		logger, _ = Initialize()
	}
	return logger
}

// InitializeLogging initializes all logging components
func InitializeLogging() error {
	// Initialize legacy logger for backward compatibility
	_, err := Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize legacy logger: %v", err)
	}

	// Pre-initialize all logger types to ensure separate log files
	GetErrorLogger()
	GetWarnLogger()
	GetInfoLogger()
	GetDebugLogger()
	GetPerformanceLogger()

	// Log successful initialization
	Info("Logging system initialized successfully")

	return nil
}
