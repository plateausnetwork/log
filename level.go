package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Level represents a distinct logging level.
type Level uint32

// The different logging levels. Note that this is basically a wrapper of the actual logrus.Level,
// just to provide a more strict ("limited", if you will) interface and to fake a 'log off'
// option.
//
// This also contains helpers to parse the value from/to string.
const (
	LevelOff Level = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

func (level Level) String() string {
	if level == LevelOff {
		return "off"
	}

	return level.toLogrus().String()
}

func (level Level) toLogrus() logrus.Level {
	switch level {
	case LevelDebug:
		return logrus.DebugLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelError:
		return logrus.ErrorLevel
	default:
		return logrus.PanicLevel // lowest level available == 'off'
	}
}

func fromLogrus(level logrus.Level) Level {
	switch level {
	case logrus.DebugLevel:
		return LevelDebug
	case logrus.InfoLevel:
		return LevelInfo
	case logrus.WarnLevel:
		return LevelWarn
	case logrus.ErrorLevel:
		return LevelError
	default:
		return LevelOff
	}
}

// ParseLevel converts the supplied string to a valid Level value.
func ParseLevel(level string) (Level, error) {
	if level == "off" {
		return LevelOff, nil
	}

	result, err := logrus.ParseLevel(level)
	if err != nil {
		return LevelOff, fmt.Errorf("unsupported log level: '%s'", level)
	}

	switch result {
	case logrus.DebugLevel:
		return LevelDebug, nil
	case logrus.InfoLevel:
		return LevelInfo, nil
	case logrus.WarnLevel:
		return LevelWarn, nil
	case logrus.ErrorLevel:
		return LevelError, nil
	default:
		return LevelOff, fmt.Errorf("unsupported log level: '%s'", level)
	}
}
