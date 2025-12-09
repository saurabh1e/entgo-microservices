package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *logrus.Logger
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `json:"level"`
	LogDir     string `json:"log_dir"`
	MaxSize    int    `json:"max_size"`    // megabytes
	MaxBackups int    `json:"max_backups"` // number of files
	MaxAge     int    `json:"max_age"`     // days
	Compress   bool   `json:"compress"`
}

// InitLogger initializes the logger with rotating file handler
func InitLogger(config LogConfig) error {
	Logger = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return err
	}

	// Set up rotating file logger for all logs
	allLogsFile := &lumberjack.Logger{
		Filename:   filepath.Join(config.LogDir, "app.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Set up rotating file logger for error logs only
	errorLogsFile := &lumberjack.Logger{
		Filename:   filepath.Join(config.LogDir, "error.log"),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Create a multi-writer for both file and console output
	Logger.SetOutput(io.MultiWriter(os.Stdout, allLogsFile))

	// Set JSON formatter for structured logging
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})

	// Hook for error logs to separate file
	Logger.AddHook(&ErrorFileHook{errorWriter: errorLogsFile})

	return nil
}

// ErrorFileHook is a hook that writes error logs to a separate file
type ErrorFileHook struct {
	errorWriter io.Writer
}

func (hook *ErrorFileHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= logrus.ErrorLevel {
		line, err := entry.String()
		if err != nil {
			return err
		}
		_, err = hook.errorWriter.Write([]byte(line))
		return err
	}
	return nil
}

func (hook *ErrorFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
}

// Convenience functions for different log levels
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// WithFields creates an entry with fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithField creates an entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithError creates an entry from the logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}
