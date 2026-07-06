// Package logger provides a small structured logger used throughout
// BlueWhale. It always writes full detail to the scan's log file, and
// additionally echoes to stdout when verbose mode is enabled.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level identifies the severity of a log entry.
type Level int

const (
	LevelInfo Level = iota
	LevelOK
	LevelWarn
	LevelError
)

func (l Level) tag() string {
	switch l {
	case LevelOK:
		return "OK"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// Logger writes timestamped entries to a log file and, optionally, mirrors
// them to stdout. It is safe for concurrent use by multiple goroutines.
type Logger struct {
	mu      sync.Mutex
	file    *os.File
	verbose bool
	fileLog *log.Logger
}

// New creates a Logger that writes to logFilePath. If verbose is true,
// entries are also printed to stdout in real time.
func New(logFilePath string, verbose bool) (*Logger, error) {
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file %q: %w", logFilePath, err)
	}

	return &Logger{
		file:    f,
		verbose: verbose,
		fileLog: log.New(f, "", 0),
	}, nil
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *Logger) write(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] [%s] %s", timestamp, level.tag(), msg)

	l.fileLog.Println(line)

	if l.verbose {
		fmt.Fprintf(consoleWriter, "[%s] %s\n", level.tag(), msg)
	}
}

// consoleWriter is a package-level indirection so tests (or future
// customization) could redirect console output; defaults to stdout.
var consoleWriter io.Writer = os.Stdout

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) { l.write(LevelInfo, format, args...) }

// OK logs a success message.
func (l *Logger) OK(format string, args ...interface{}) { l.write(LevelOK, format, args...) }

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) { l.write(LevelWarn, format, args...) }

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) { l.write(LevelError, format, args...) }
