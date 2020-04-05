package log

import (
	"fmt"

	logrus "github.com/sirupsen/logrus"
)

// F represents a key-value list of custom fields to be added to the log entry.
// For more information, take a look at the With function.
type F = logrus.Fields

// With returns a new log entry with the supplied key-value fields. Note that
// this function does not register the entry in the log; it only creates a new entry.
// To actually register the information on the logger use the Debug, Info, Warn or Error
// methods.
func With(fields F) *Entry {
	loggerLock.RLock()
	defer loggerLock.RUnlock()
	return &Entry{inner: logger.WithFields(fields)}
}

// WithError returns a new log entry for error. The entry is created by populating the fields
// 'error' and 'stack' and, like the entry created by With, it is not immediately registered in
// the logger.
//
// The big difference between With and With error - apart from extracting the error information - is
// that WithError returns an entry that is 'locked' in a 'error state', so the only way to actually
// register the entry is calling Error (the other methods - Debug, Info, etc. - are not available).
func WithError(err error) *ErrorEntry {
	return &ErrorEntry{inner: logger.WithFields(fieldsFromError(err))}
}

// PrintError is an auxiliary function to display user-friendly messages on stdout. PrintError
// registers the supplied error on the logger and prints the user-friendly message to stdout, but
// only if the stdout logger is turned off, to avoid shoing 'duplicated' error messages to the user.
func PrintError(err error, message string) {
	loggerLock.RLock()
	defer loggerLock.RUnlock()

	logger.WithFields(fieldsFromError(err)).Error()

	if stdoutHook.level == LevelOff.toLogrus() {
		stdoutHook.writer.Write([]byte(fmt.Sprintf("%s\n", message))) // nolint: errcheck
	}
}

// Debug registers a log entry in the 'Debug' level.
func Debug(message string) {
	logger.Debug(message)
}

// Info registers a log entry in the 'Info' level.
func Info(message string) {
	logger.Info(message)
}

// Warn registers a log entry in the 'Warn' level.
func Warn(message string) {
	logger.Warn(message)
}

// Error registers a log entry in the 'Error' level. The error stack trace is also
// recorded.
func Error(err error) {
	logger.WithFields(fieldsFromError(err)).Error()
}
