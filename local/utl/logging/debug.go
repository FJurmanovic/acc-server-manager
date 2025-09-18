package logging

import (
	"fmt"
	"runtime"
	"sync"
)

type DebugLogger struct {
	base *BaseLogger
}

func NewDebugLogger() *DebugLogger {
	base, _ := InitializeBase("debug")
	return &DebugLogger{
		base: base,
	}
}

func (dl *DebugLogger) Log(format string, v ...interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, format, v...)
	}
}

func (dl *DebugLogger) LogWithContext(context string, format string, v ...interface{}) {
	if dl.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		dl.base.Log(LogLevelDebug, contextualFormat, v...)
	}
}

func (dl *DebugLogger) LogFunction(functionName string, args ...interface{}) {
	if dl.base != nil {
		if len(args) > 0 {
			dl.base.Log(LogLevelDebug, "FUNCTION [%s] called with args: %+v", functionName, args)
		} else {
			dl.base.Log(LogLevelDebug, "FUNCTION [%s] called", functionName)
		}
	}
}

func (dl *DebugLogger) LogVariable(varName string, value interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "VARIABLE [%s]: %+v", varName, value)
	}
}

func (dl *DebugLogger) LogState(component string, state interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "STATE [%s]: %+v", component, state)
	}
}

func (dl *DebugLogger) LogSQL(query string, args ...interface{}) {
	if dl.base != nil {
		if len(args) > 0 {
			dl.base.Log(LogLevelDebug, "SQL: %s | Args: %+v", query, args)
		} else {
			dl.base.Log(LogLevelDebug, "SQL: %s", query)
		}
	}
}

func (dl *DebugLogger) LogMemory() {
	if dl.base != nil {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		dl.base.Log(LogLevelDebug, "MEMORY: Alloc = %d KB, TotalAlloc = %d KB, Sys = %d KB, NumGC = %d",
			bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
	}
}

func (dl *DebugLogger) LogGoroutines() {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "GOROUTINES: %d active", runtime.NumGoroutine())
	}
}

func (dl *DebugLogger) LogTiming(operation string, duration interface{}) {
	if dl.base != nil {
		dl.base.Log(LogLevelDebug, "TIMING [%s]: %v", operation, duration)
	}
}

func bToKb(b uint64) uint64 {
	return b / 1024
}

var (
	debugLogger *DebugLogger
	debugOnce   sync.Once
)

func GetDebugLogger() *DebugLogger {
	debugOnce.Do(func() {
		debugLogger = NewDebugLogger()
	})
	return debugLogger
}

func Debug(format string, v ...interface{}) {
	GetDebugLogger().Log(format, v...)
}

func DebugWithContext(context string, format string, v ...interface{}) {
	GetDebugLogger().LogWithContext(context, format, v...)
}

func DebugFunction(functionName string, args ...interface{}) {
	GetDebugLogger().LogFunction(functionName, args...)
}

func DebugVariable(varName string, value interface{}) {
	GetDebugLogger().LogVariable(varName, value)
}

func DebugState(component string, state interface{}) {
	GetDebugLogger().LogState(component, state)
}

func DebugSQL(query string, args ...interface{}) {
	GetDebugLogger().LogSQL(query, args...)
}

func DebugMemory() {
	GetDebugLogger().LogMemory()
}

func DebugGoroutines() {
	GetDebugLogger().LogGoroutines()
}

func DebugTiming(operation string, duration interface{}) {
	GetDebugLogger().LogTiming(operation, duration)
}
