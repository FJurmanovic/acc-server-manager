package logging

import (
	"fmt"
)

// WarnLogger handles warn-level logging
type WarnLogger struct {
	base *BaseLogger
}

// NewWarnLogger creates a new warn logger instance
func NewWarnLogger() *WarnLogger {
	return &WarnLogger{
		base: GetBaseLogger("warn"),
	}
}

// Log writes a warn-level log entry
func (wl *WarnLogger) Log(format string, v ...interface{}) {
	if wl.base != nil {
		wl.base.Log(LogLevelWarn, format, v...)
	}
}

// LogWithContext writes a warn-level log entry with additional context
func (wl *WarnLogger) LogWithContext(context string, format string, v ...interface{}) {
	if wl.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		wl.base.Log(LogLevelWarn, contextualFormat, v...)
	}
}

// LogDeprecation logs a deprecation warning
func (wl *WarnLogger) LogDeprecation(feature string, alternative string) {
	if wl.base != nil {
		if alternative != "" {
			wl.base.Log(LogLevelWarn, "DEPRECATED: %s is deprecated, use %s instead", feature, alternative)
		} else {
			wl.base.Log(LogLevelWarn, "DEPRECATED: %s is deprecated", feature)
		}
	}
}

// LogConfiguration logs configuration-related warnings
func (wl *WarnLogger) LogConfiguration(setting string, message string) {
	if wl.base != nil {
		wl.base.Log(LogLevelWarn, "CONFIG WARNING [%s]: %s", setting, message)
	}
}

// LogPerformance logs performance-related warnings
func (wl *WarnLogger) LogPerformance(operation string, threshold string, actual string) {
	if wl.base != nil {
		wl.base.Log(LogLevelWarn, "PERFORMANCE WARNING [%s]: exceeded threshold %s, actual: %s", operation, threshold, actual)
	}
}

// Global warn logger instance
var warnLogger *WarnLogger

// GetWarnLogger returns the global warn logger instance
func GetWarnLogger() *WarnLogger {
	if warnLogger == nil {
		warnLogger = NewWarnLogger()
	}
	return warnLogger
}

// Warn logs a warn-level message using the global warn logger
func Warn(format string, v ...interface{}) {
	GetWarnLogger().Log(format, v...)
}

// WarnWithContext logs a warn-level message with context using the global warn logger
func WarnWithContext(context string, format string, v ...interface{}) {
	GetWarnLogger().LogWithContext(context, format, v...)
}

// WarnDeprecation logs a deprecation warning using the global warn logger
func WarnDeprecation(feature string, alternative string) {
	GetWarnLogger().LogDeprecation(feature, alternative)
}

// WarnConfiguration logs configuration-related warnings using the global warn logger
func WarnConfiguration(setting string, message string) {
	GetWarnLogger().LogConfiguration(setting, message)
}

// WarnPerformance logs performance-related warnings using the global warn logger
func WarnPerformance(operation string, threshold string, actual string) {
	GetWarnLogger().LogPerformance(operation, threshold, actual)
}
