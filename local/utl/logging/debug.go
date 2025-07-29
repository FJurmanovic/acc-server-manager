package logging

import (
	"fmt"
	"runtime"
	"sync"
)

// DebugLogger handles debug-level logging
type DebugLogger struct {
	base *BaseLogger
}

// NewDebugLogger creates a new debug logger instance
func NewDebugLogger() *DebugLogger {
	base, _ := InitializeBase("debug")
	return &DebugLogger{
		base: base,
	}
}

// Log writes a debug-level log entry
func (dl *DebugLogger) Log(format string, v ...interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, format, v...)
	}
}

// LogWithContext writes a debug-level log entry with additional context
func (dl *DebugLogger) LogWithContext(context string, format string, v ...interface{}) {
	if dl.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		dl.base.Log(LogLevelDebug, contextualFormat, v...)
	}
}

// LogFunction logs function entry and exit for debugging
func (dl *DebugLogger) LogFunction(functionName string, args ...interface{}) {
	if dl.base != nil {
		if len(args) > 0 {
			dl.base.Log(LogLevelDebug, "FUNCTION [%s] called with args: %+v", functionName, args)
		} else {
			dl.base.Log(LogLevelDebug, "FUNCTION [%s] called", functionName)
		}
	}
}

// LogVariable logs variable values for debugging
func (dl *DebugLogger) LogVariable(varName string, value interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "VARIABLE [%s]: %+v", varName, value)
	}
}

// LogState logs application state information
func (dl *DebugLogger) LogState(component string, state interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "STATE [%s]: %+v", component, state)
	}
}

// LogSQL logs SQL queries for debugging
func (dl *DebugLogger) LogSQL(query string, args ...interface{}) {
	if dl.base != nil {
		if len(args) > 0 {
			dl.base.Log(LogLevelDebug, "SQL: %s | Args: %+v", query, args)
		} else {
			dl.base.Log(LogLevelDebug, "SQL: %s", query)
		}
	}
}

// LogMemory logs memory usage information
func (dl *DebugLogger) LogMemory() {
	if dl.base != nil {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		dl.base.Log(LogLevelDebug, "MEMORY: Alloc = %d KB, TotalAlloc = %d KB, Sys = %d KB, NumGC = %d",
			bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
	}
}

// LogGoroutines logs current number of goroutines
func (dl *DebugLogger) LogGoroutines() {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "GOROUTINES: %d active", runtime.NumGoroutine())
	}
}

// LogTiming logs timing information for performance debugging
func (dl *DebugLogger) LogTiming(operation string, duration interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "TIMING [%s]: %v", operation, duration)
	}
}

// Helper function to convert bytes to kilobytes
func bToKb(b uint64) uint64 {
	return b / 1024
}

// Global debug logger instance
var (
	debugLogger *DebugLogger
	debugOnce   sync.Once
)

// GetDebugLogger returns the global debug logger instance
func GetDebugLogger() *DebugLogger {
	debugOnce.Do(func() {
		debugLogger = NewDebugLogger()
	})
	return debugLogger
}

// Debug logs a debug-level message using the global debug logger
func Debug(format string, v ...interface{}) {
	GetDebugLogger().Log(format, v...)
}

// DebugWithContext logs a debug-level message with context using the global debug logger
func DebugWithContext(context string, format string, v ...interface{}) {
	GetDebugLogger().LogWithContext(context, format, v...)
}

// DebugFunction logs function entry and exit using the global debug logger
func DebugFunction(functionName string, args ...interface{}) {
	GetDebugLogger().LogFunction(functionName, args...)
}

// DebugVariable logs variable values using the global debug logger
func DebugVariable(varName string, value interface{}) {
	GetDebugLogger().LogVariable(varName, value)
}

// DebugState logs application state information using the global debug logger
func DebugState(component string, state interface{}) {
	GetDebugLogger().LogState(component, state)
}

// DebugSQL logs SQL queries using the global debug logger
func DebugSQL(query string, args ...interface{}) {
	GetDebugLogger().LogSQL(query, args...)
}

// DebugMemory logs memory usage information using the global debug logger
func DebugMemory() {
	GetDebugLogger().LogMemory()
}

// DebugGoroutines logs current number of goroutines using the global debug logger
func DebugGoroutines() {
	GetDebugLogger().LogGoroutines()
}

// DebugTiming logs timing information using the global debug logger
func DebugTiming(operation string, duration interface{}) {
	GetDebugLogger().LogTiming(operation, duration)
}
