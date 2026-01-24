package logger

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init() {
	Log = logrus.New()

	// Set JSON formatter for structured logging
	Log.SetFormatter(&logrus.JSONFormatter{})

	// Set log level (can be configured via env var later)
	Log.SetLevel(logrus.InfoLevel)
}

// Get returns the logger instance, initializing it if needed
func Get() *logrus.Logger {
	if Log == nil {
		Init()
	}
	return Log
}

// Convenience wrapper functions
func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

func WithError(err error) *logrus.Entry {
	return Get().WithError(err)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}
