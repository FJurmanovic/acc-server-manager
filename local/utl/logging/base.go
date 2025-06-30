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
	baseOnce   sync.Once
	baseLogger *BaseLogger
)

// BaseLogger provides the core logging functionality
type BaseLogger struct {
	file        *os.File
	logger      *log.Logger
	mu          sync.RWMutex
	initialized bool
}

// LogLevel represents different logging levels
type LogLevel string

const (
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelPanic LogLevel = "PANIC"
)

// Initialize creates or gets the singleton base logger instance
func InitializeBase(tp string) (*BaseLogger, error) {
	var err error
	baseOnce.Do(func() {
		baseLogger, err = newBaseLogger(tp)
	})
	return baseLogger, err
}

func newBaseLogger(tp string) (*BaseLogger, error) {
	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Open log file with date in name
	logPath := filepath.Join("logs", fmt.Sprintf("acc-server-%s-%s.log", time.Now().Format("2006-01-02"), tp))
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both file and console
	multiWriter := io.MultiWriter(file, os.Stdout)

	// Create base logger
	logger := &BaseLogger{
		file:        file,
		logger:      log.New(multiWriter, "", 0),
		initialized: true,
	}

	return logger, nil
}

// GetBaseLogger returns the singleton base logger instance
func GetBaseLogger(tp string) *BaseLogger {
	if baseLogger == nil {
		baseLogger, _ = InitializeBase(tp)
	}
	return baseLogger
}

// Close closes the log file
func (bl *BaseLogger) Close() error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if bl.file != nil {
		return bl.file.Close()
	}
	return nil
}

// Log writes a log entry with the specified level
func (bl *BaseLogger) Log(level LogLevel, format string, v ...interface{}) {
	if bl == nil || !bl.initialized {
		return
	}

	bl.mu.RLock()
	defer bl.mu.RUnlock()

	// Get caller info (skip 2 frames: this function and the calling Log function)
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)

	// Format message
	msg := fmt.Sprintf(format, v...)

	// Format final log line
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s",
		time.Now().Format(timeFormat),
		string(level),
		file,
		line,
		msg,
	)

	bl.logger.Println(logLine)
}

// LogWithCaller writes a log entry with custom caller depth
func (bl *BaseLogger) LogWithCaller(level LogLevel, callerDepth int, format string, v ...interface{}) {
	if bl == nil || !bl.initialized {
		return
	}

	bl.mu.RLock()
	defer bl.mu.RUnlock()

	// Get caller info with custom depth
	_, file, line, _ := runtime.Caller(callerDepth)
	file = filepath.Base(file)

	// Format message
	msg := fmt.Sprintf(format, v...)

	// Format final log line
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s",
		time.Now().Format(timeFormat),
		string(level),
		file,
		line,
		msg,
	)

	bl.logger.Println(logLine)
}

// IsInitialized returns whether the base logger is initialized
func (bl *BaseLogger) IsInitialized() bool {
	if bl == nil {
		return false
	}
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.initialized
}

// RecoverAndLog recovers from panics and logs them
func RecoverAndLog() {
	baseLogger := GetBaseLogger("log")
	if baseLogger != nil && baseLogger.IsInitialized() {
		if r := recover(); r != nil {
			// Get stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			baseLogger.LogWithCaller(LogLevelPanic, 2, "Recovered from panic: %v\nStack Trace:\n%s", r, stackTrace)

			// Re-panic to maintain original behavior if needed
			panic(r)
		}
	}
}
