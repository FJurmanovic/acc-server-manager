package logging

import (
	"fmt"
	"sync"
)

type WarnLogger struct {
	base *BaseLogger
}

func NewWarnLogger() *WarnLogger {
	base, _ := InitializeBase("warn")
	return &WarnLogger{
		base: base,
	}
}

func (wl *WarnLogger) Log(format string, v ...interface{}) {
	if wl.base != nil {
		wl.base.Log(LogLevelWarn, format, v...)
	}
}

func (wl *WarnLogger) LogWithContext(context string, format string, v ...interface{}) {
	if wl.base != nil {
		contextualFormat := fmt.Sprintf("[%s] %s", context, format)
		wl.base.Log(LogLevelWarn, contextualFormat, v...)
	}
}

func (wl *WarnLogger) LogDeprecation(feature string, alternative string) {
	if wl.base != nil {
		if alternative != "" {
			wl.base.Log(LogLevelWarn, "DEPRECATED: %s is deprecated, use %s instead", feature, alternative)
		} else {
			wl.base.Log(LogLevelWarn, "DEPRECATED: %s is deprecated", feature)
		}
	}
}

func (wl *WarnLogger) LogConfiguration(setting string, message string) {
	if wl.base != nil {
		wl.base.Log(LogLevelWarn, "CONFIG WARNING [%s]: %s", setting, message)
	}
}

func (wl *WarnLogger) LogPerformance(operation string, threshold string, actual string) {
	if wl.base != nil {
		wl.base.Log(LogLevelWarn, "PERFORMANCE WARNING [%s]: exceeded threshold %s, actual: %s", operation, threshold, actual)
	}
}

var (
	warnLogger *WarnLogger
	warnOnce   sync.Once
)

func GetWarnLogger() *WarnLogger {
	warnOnce.Do(func() {
		warnLogger = NewWarnLogger()
	})
	return warnLogger
}

func Warn(format string, v ...interface{}) {
	GetWarnLogger().Log(format, v...)
}

func WarnWithContext(context string, format string, v ...interface{}) {
	GetWarnLogger().LogWithContext(context, format, v...)
}

func WarnDeprecation(feature string, alternative string) {
	GetWarnLogger().LogDeprecation(feature, alternative)
}

func WarnConfiguration(setting string, message string) {
	GetWarnLogger().LogConfiguration(setting, message)
}

func WarnPerformance(operation string, threshold string, actual string) {
	GetWarnLogger().LogPerformance(operation, threshold, actual)
}
