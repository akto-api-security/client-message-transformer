package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel defines the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging with levels
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger with specified level
func NewLogger(levelStr string) *Logger {
	level := INFO
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	}

	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// formatMessage creates a formatted log message
func (l *Logger) formatMessage(levelStr string, msg string) string {
	return fmt.Sprintf("[%s] %s | %s", time.Now().Format("2006-01-02 15:04:05"), levelStr, msg)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	if l.level <= DEBUG {
		l.logger.Println(l.formatMessage("DEBUG", msg))
	}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	if l.level <= INFO {
		l.logger.Println(l.formatMessage("INFO ", msg))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	if l.level <= WARN {
		l.logger.Println(l.formatMessage("WARN ", msg))
	}
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	if l.level <= ERROR {
		l.logger.Println(l.formatMessage("ERROR", msg))
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}
