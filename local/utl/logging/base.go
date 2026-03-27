package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	timeFormat = "2006-01-02 15:04:05.000"
)

type BaseLogger struct {
	file        *os.File
	logger      *log.Logger
	mu          sync.RWMutex
	initialized bool
}

type LogLevel string

const (
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelPanic LogLevel = "PANIC"
)

func InitializeBase(tp string) (*BaseLogger, error) {
	return newBaseLogger(tp)
}

func newBaseLogger(tp string) (*BaseLogger, error) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	logPath := filepath.Join("logs", fmt.Sprintf("acc-server-%s-%s.log", time.Now().Format("2006-01-02"), tp))
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)

	logger := &BaseLogger{
		file:        file,
		logger:      log.New(multiWriter, "", 0),
		initialized: true,
	}

	return logger, nil
}

func GetBaseLogger(tp string) *BaseLogger {
	baseLogger, _ := InitializeBase(tp)
	return baseLogger
}

func (bl *BaseLogger) Close() error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if bl.file != nil {
		return bl.file.Close()
	}
	return nil
}

func (bl *BaseLogger) Log(level LogLevel, format string, v ...interface{}) {
	if bl == nil || !bl.initialized {
		return
	}

	bl.mu.RLock()
	defer bl.mu.RUnlock()

	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)

	msg := fmt.Sprintf(format, v...)

	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s",
		time.Now().Format(timeFormat),
		string(level),
		file,
		line,
		msg,
	)

	bl.logger.Println(logLine)
}

func (bl *BaseLogger) LogWithCaller(level LogLevel, callerDepth int, format string, v ...interface{}) {
	if bl == nil || !bl.initialized {
		return
	}

	bl.mu.RLock()
	defer bl.mu.RUnlock()

	_, file, line, _ := runtime.Caller(callerDepth)
	file = filepath.Base(file)

	msg := fmt.Sprintf(format, v...)

	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s",
		time.Now().Format(timeFormat),
		string(level),
		file,
		line,
		msg,
	)

	bl.logger.Println(logLine)
}

func (bl *BaseLogger) IsInitialized() bool {
	if bl == nil {
		return false
	}
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.initialized
}

func RecoverAndLog() {
	baseLogger := GetBaseLogger("panic")
	if baseLogger != nil && baseLogger.IsInitialized() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			baseLogger.LogWithCaller(LogLevelPanic, 2, "Recovered from panic: %v\nStack Trace:\n%s", r, stackTrace)

			panic(r)
		}
	}
}
