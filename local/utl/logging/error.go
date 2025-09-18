package logging

import (
	"fmt"
	"runtime"
	"sync"
)

type ErrorLogger struct {
	base *BaseLogger
}

func NewErrorLogger() *ErrorLogger {
	base, _ := InitializeBase("error")
	return &ErrorLogger{
		base: base,
	}
}

func (el *ErrorLogger) Log(format string, v ...interface{}) {
	if el.base != nil {
		el.base.Log(LogLevelError, format, v...)
	}
}

func (el *ErrorLogger) LogWithContext(context string, format string, v ...interface{}) {
	if el.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		el.base.Log(LogLevelError, contextualFormat, v...)
	}
}

func (el *ErrorLogger) LogError(err error, message ...string) {
	if el.base != nil && err != nil {
		if len(message) > 0 {
			el.base.Log(LogLevelError, "%s: %v", message[0], err)
		} else {
			el.base.Log(LogLevelError, "Error: %v", err)
		}
	}
}

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

func (el *ErrorLogger) LogFatal(format string, v ...interface{}) {
	if el.base != nil {
		el.base.Log(LogLevelError, "[FATAL] "+format, v...)
		panic(fmt.Sprintf(format, v...))
	}
}

var (
	errorLogger *ErrorLogger
	errorOnce   sync.Once
)

func GetErrorLogger() *ErrorLogger {
	errorOnce.Do(func() {
		errorLogger = NewErrorLogger()
	})
	return errorLogger
}

func Error(format string, v ...interface{}) {
	GetErrorLogger().Log(format, v...)
}

func ErrorWithContext(context string, format string, v ...interface{}) {
	GetErrorLogger().LogWithContext(context, format, v...)
}

func LogError(err error, message ...string) {
	GetErrorLogger().LogError(err, message...)
}

func ErrorWithStackTrace(format string, v ...interface{}) {
	GetErrorLogger().LogWithStackTrace(format, v...)
}

func Fatal(format string, v ...interface{}) {
	GetErrorLogger().LogFatal(format, v...)
}
