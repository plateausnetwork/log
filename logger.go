package log

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"

	"github.com/rhizomplatform/fs"
)

var (
	loggerLock sync.RWMutex
	logger     *logrus.Logger
	fileHook   *levelWriterHook
	stdoutHook *levelWriterHook
)

// Setup configures and starts a new global logger instance. If the global logger is
// already configured, the call is ignored.
func Setup(logPath fs.Path, logsufix string, purgeMinutes, rotateMinutes int) {
	if logger != nil {
		return
	}

	loggerLock.Lock()
	defer loggerLock.Unlock()

	if err := logPath.MkdirAll(); err != nil {
		panic(err)
	}

	path := logPath.Join(logsufix + ".log").String()

	rotate, err := rotatelogs.New(
		strings.Replace(path, logsufix+".log", "%Y%m%d%H%M-"+logsufix+".json", -1),
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(purgeMinutes)*time.Minute),
		rotatelogs.WithRotationTime(time.Duration(rotateMinutes)*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	logger = logrus.New()

	// Hooks to control where/what will be logged on
	fileHook = &levelWriterHook{
		level:          logrus.InfoLevel,
		writer:         rotate,
		formatter:      &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano},
		showErrorStack: true,
	}

	stdoutHook = &levelWriterHook{
		level:     logrus.InfoLevel,
		writer:    os.Stdout,
		formatter: &logrus.TextFormatter{ForceColors: true},
	}

	logger.AddHook(fileHook)
	logger.AddHook(stdoutHook)

	// Will always discard by default, since we're controlling this
	// inside our hooks
	logger.Out = ioutil.Discard

	// Will always be Debug by default, since the actual control is
	// made by the hooks
	logger.SetLevel(logrus.DebugLevel)
}

// TearDown disables the global logger, undoing the configuration steps made
// in the Setup function.
func TearDown() {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	logger = nil
	fileHook = nil
	stdoutHook = nil
}

// GetStdoutLevel returns the current log level of stdout
func GetStdoutLevel() Level {
	loggerLock.RLock()
	defer loggerLock.RUnlock()

	return fromLogrus(stdoutHook.level)
}

// SetStdoutLevel configures the log level used on the stdout output. The
// default initial level is 'Info'.
func SetStdoutLevel(level Level) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	stdoutHook.level = level.toLogrus()
}

// SetFileLevel configures the log level used on the file output. The
// default initial level is 'Info'.
func SetFileLevel(level Level) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	fileHook.level = level.toLogrus()
}

// RedirectStdout redirects the stdout logger output to the supplied Writer.
// This function is only useful for testing purposes, so do not use this
// to turn off the logger; if you want to disable the stdout logger use
// SetLevel(LevelOff) instead.
func RedirectStdout(target io.Writer) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	stdoutHook.writer = target
}

// RestoreStdout restore the stdout redirection made by RedirectStdout. Again, this
// function is useful only for testing puroses.
func RestoreStdout() {
	if stdoutHook.writer == os.Stdout {
		return
	}

	loggerLock.Lock()
	defer loggerLock.Unlock()

	stdoutHook.writer = os.Stdout
}
