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
	logger     *Logger
	once       sync.Once
	timeFormat = "2006-01-02 15:04:05.000"
)

type Logger struct {
	file   *os.File
	logger *log.Logger
}

// Initialize creates or gets the singleton logger instance
func Initialize() (*Logger, error) {
	var err error
	once.Do(func() {
		logger, err = newLogger()
	})
	return logger, err
}

func newLogger() (*Logger, error) {
	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Open log file with date in name
	logPath := filepath.Join("logs", fmt.Sprintf("acc-server-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both file and console
	multiWriter := io.MultiWriter(file, os.Stdout)
	
	// Create logger with custom prefix
	logger := &Logger{
		file:   file,
		logger: log.New(multiWriter, "", 0),
	}

	return logger, nil
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) log(level, format string, v ...interface{}) {
	// Get caller info
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)

	// Format message
	msg := fmt.Sprintf(format, v...)
	
	// Format final log line
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s",
		time.Now().Format(timeFormat),
		level,
		file,
		line,
		msg,
	)

	l.logger.Println(logLine)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log("INFO", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log("ERROR", format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log("WARN", format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log("DEBUG", format, v...)
}

func (l *Logger) Panic(format string) {
	l.Panic("PANIC " + format)
}

// Global convenience functions
func Info(format string, v ...interface{}) {
	if logger != nil {
		logger.Info(format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if logger != nil {
		logger.Error(format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if logger != nil {
		logger.Warn(format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if logger != nil {
		logger.Debug(format, v...)
	}
}

func Panic(format string) {
	if logger != nil {
		logger.Panic(format)
	}
}

// RecoverAndLog recovers from panics and logs them
func RecoverAndLog() {
	if r := recover(); r != nil {
		if logger != nil {
			// Get stack trace
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			logger.log("PANIC", "Recovered from panic: %v\nStack Trace:\n%s", r, stackTrace)
		}
	}
} 