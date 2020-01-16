package log

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type levelWriterHook struct {
	level          logrus.Level
	writer         io.Writer
	formatter      logrus.Formatter
	showErrorStack bool
}

func (hook *levelWriterHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *levelWriterHook) Fire(entry *logrus.Entry) error {
	loggerLock.RLock()
	defer loggerLock.RUnlock()

	if entry.Level > hook.level {
		return nil
	}

	if entry.Level == logrus.ErrorLevel {
		errMsg, stack := hook.extractError(entry)

		// replace the error struct with the actual message
		if errMsg != "" {
			entry.Data["error"] = errMsg
		}

		// fille the entry message with a default value
		if entry.Message == "" && errMsg != "" {
			entry.Message = errMsg
		}

		// avoid showing the same thing twice
		if entry.Message == errMsg {
			delete(entry.Data, "error")
		}

		// if the hook does not print stacks, we temporarily
		// remove this information
		if stack != "" && !hook.showErrorStack {
			delete(entry.Data, "stack")
			defer func() {
				entry.Data["stack"] = stack
			}()
		}
	}

	msg, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = hook.writer.Write(msg)
	return err
}

func (hook *levelWriterHook) extractError(entry *logrus.Entry) (string, string) {
	var errMsg, stack string

	if temp, ok := entry.Data["error"]; ok {
		switch v := temp.(type) {
		case error:
			errMsg = v.Error()
		case fmt.Stringer:
			errMsg = v.String()
		case string:
			errMsg = v
		}
	}

	if temp, ok := entry.Data["stack"]; ok {
		stack, _ = temp.(string)
	}

	return errMsg, stack
}
