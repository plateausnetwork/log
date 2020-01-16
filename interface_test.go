package log_test

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"path"
	"strings"
	"testing"

	pkgerr "github.com/pkg/errors"
	"github.com/rhizomplatform/fs"
	"github.com/rhizomplatform/log"
)

type outputTest struct {
	writeFn func(string)
	content string
	file    bool
	screen  bool
}

// used to check a handful of different output scenarios
func outputTestRunner(t *testing.T, group string, tests []outputTest, fileLevel, stdoutLevel log.Level) {
	baseFolder, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("error creating temp directory:", err)
	}
	defer fs.RemoveAll(baseFolder)

	log.Setup(fs.Path(baseFolder), "mysufix", 2, 1)
	defer log.TearDown()

	var buffer bytes.Buffer
	screen := bufio.NewWriter(&buffer)

	log.SetFileLevel(fileLevel)
	log.SetStdoutLevel(stdoutLevel)
	log.RedirectStdout(screen)
	defer log.RestoreStdout()

	for _, test := range tests {
		test.writeFn(test.content)
	}

	screen.Flush()
	screenContent := buffer.String()

	logfile := fs.Path(baseFolder).Join("mysufix.log")

	// small trick to make sure the file exists even no log is written
	if !logfile.FileExists() {
		if f, err := logfile.Create(); err != nil {
			t.Errorf("error creating file '%s': %v", logfile, err)
		} else {
			f.Close()
		}
	}

	b, err := logfile.ReadAll()
	if err != nil {
		t.Errorf("error reading file '%s': %v", logfile, err)
	}
	logContent := string(b)

	for i, test := range tests {
		fileExist := strings.Contains(logContent, test.content)
		screenExist := strings.Contains(screenContent, test.content)

		if fileExist && !test.file {
			t.Errorf("[Group: %s] Case %d, log line '%s' should not exist on file", group, i, test.content)
		} else if !fileExist && test.file {
			t.Errorf("[Group: %s] Case %d, log line '%s' should exist on file", group, i, test.content)
		}

		if screenExist && !test.screen {
			t.Errorf("[Group: %s] Case %d, log line '%s' should not exist on screen", group, i, test.content)
		} else if !screenExist && test.screen {
			t.Errorf("[Group: %s] Case %d, log line '%s' should exist on screen", group, i, test.content)
		}
	}
}

// used to easily handle the log setup/collection boilerpart
func collectLog(t *testing.T, handler func()) (string, string) {
	baseFolder, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("error creating temp directory:", err)
	}
	defer fs.RemoveAll(baseFolder)

	log.Setup(fs.Path(baseFolder), "mysufix", 2, 1)
	defer log.TearDown()

	log.SetFileLevel(log.LevelDebug)
	log.SetStdoutLevel(log.LevelDebug)

	var buffer bytes.Buffer
	screen := bufio.NewWriter(&buffer)

	log.RedirectStdout(screen)
	defer log.RestoreStdout()

	handler()

	screen.Flush()

	logfile := path.Join(baseFolder, "mysufix.log")
	b, err := fs.ReadAll(logfile)
	if err != nil {
		t.Errorf("error reading file '%s': %v", logfile, err)
	}

	return string(b), buffer.String()
}

func splitLines(content string) []string {
	lines := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func TestLogOutput(t *testing.T) {

	// Group 1: File and Screen full log
	group1 := []outputTest{
		{writeFn: log.Debug, content: "log-debug", file: true, screen: true},
		{writeFn: log.Info, content: "log-info", file: true, screen: true},
		{writeFn: log.Warn, content: "log-warn", file: true, screen: true},
	}
	outputTestRunner(t, "G1", group1, log.LevelDebug, log.LevelDebug)

	// Group 2: Partial log on screen
	group2 := []outputTest{
		{writeFn: log.Debug, content: "log-debug", file: true, screen: false},
		{writeFn: log.Info, content: "log-info", file: true, screen: false},
		{writeFn: log.Warn, content: "log-warn", file: true, screen: true},
	}
	outputTestRunner(t, "G2", group2, log.LevelDebug, log.LevelWarn)

	// Group 3: Partial log on file
	group3 := []outputTest{
		{writeFn: log.Debug, content: "log-debug", file: false, screen: true},
		{writeFn: log.Info, content: "log-info", file: false, screen: true},
		{writeFn: log.Warn, content: "log-warn", file: true, screen: true},
	}
	outputTestRunner(t, "G3", group3, log.LevelWarn, log.LevelDebug)

	// Group 4: No logging
	group4 := []outputTest{
		{writeFn: log.Debug, content: "log-debug", file: false, screen: false},
		{writeFn: log.Info, content: "log-info", file: false, screen: false},
		{writeFn: log.Warn, content: "log-warn", file: false, screen: false},
	}
	outputTestRunner(t, "G4", group4, log.LevelError, log.LevelError)
}

func TestLogNoScreenOutput(t *testing.T) {
	errorWrapper := func(msg string) {
		log.Error(errors.New(msg))
	}

	tests := []outputTest{
		{writeFn: log.Debug, content: "log-debug", file: true, screen: false},
		{writeFn: log.Info, content: "log-info", file: true, screen: false},
		{writeFn: log.Warn, content: "log-warn", file: true, screen: false},
		{writeFn: errorWrapper, content: "log-error", file: true, screen: false},
	}
	outputTestRunner(t, "no-screen-output", tests, log.LevelDebug, log.LevelOff)
}

func TestStructuredLogging(t *testing.T) {
	tests := []struct {
		fields   log.F
		expected string
	}{
		{fields: log.F{"foo": 1}, expected: "\"foo\":1"},
		{fields: log.F{"foo": "1"}, expected: "\"foo\":\"1\""},
		{fields: log.F{"foo": 1, "bar": 2}, expected: "\"foo\":1"},
		{fields: log.F{"foo": 1, "bar": 2}, expected: "\"bar\":2"},
	}

	logContent, _ := collectLog(t, func() {
		for _, test := range tests {
			log.With(test.fields).Info("?")
		}
	})

	for i, test := range tests {
		if !strings.Contains(logContent, test.expected) {
			t.Errorf("Case %d, log line '%s' not found", i, test.expected)
		}
	}
}

// aux functions to generate a stack depth
func auxStack(one, withStack bool) error {
	if one {
		return auxStack1(withStack)
	}

	return auxStack2(withStack)
}

func auxStack1(withStack bool) error {
	return auxStackErr(withStack)
}

func auxStack2(withStack bool) error {
	return auxStackErr(withStack)
}

func auxStackErr(withStack bool) error {
	if withStack {
		return pkgerr.New("some error")
	}

	return errors.New("some error")
}

func TestLogStackFrame(t *testing.T) {
	tests := []struct {
		err      error
		stack1   bool
		stack2   bool
		stackErr bool
	}{
		{err: auxStack(false, false)},
		{err: auxStack(false, true), stack2: true, stackErr: true},
		{err: auxStack(true, false)},
		{err: auxStack(true, true), stack1: true, stackErr: true},
	}

	logContent, _ := collectLog(t, func() {
		for _, test := range tests {
			log.Error(test.err)
		}
	})

	logLines := splitLines(logContent)

	if len(logLines) != len(tests) {
		t.Errorf("Wrong number of log lines: expected '%d', received '%d'", len(tests), len(logLines))
		return
	}

	for i, test := range tests {
		line := logLines[i]
		has1 := strings.Contains(line, "Stack1")
		has2 := strings.Contains(line, "Stack2")
		hasE := strings.Contains(line, "StackErr")

		if test.stack1 && !has1 {
			t.Errorf("Case %d, auxStack1 should appear in the log stack", i)
		} else if !test.stack1 && has1 {
			t.Errorf("Case %d, auxStack1 should not appear in the log stack", i)
		}

		if test.stack2 && !has2 {
			t.Errorf("Case %d, auxStack2 should appear in the log stack", i)
		} else if !test.stack2 && has2 {
			t.Errorf("Case %d, auxStack2 should not appear in the log stack", i)
		}

		if test.stackErr && !hasE {
			t.Errorf("Case %d, auxStackErr should appear in the log stack", i)
		} else if !test.stack1 && has1 {
			t.Errorf("Case %d, auxStackErr should not appear in the log stack", i)
		}
	}
}

func TestLogNoStackOnScreen(t *testing.T) {
	const numErrors = 3

	logContent, screenContent := collectLog(t, func() {
		log.Error(errors.New("e1"))
		log.With(log.F{}).Error(errors.New("e2"))
		log.WithError(errors.New("e3")).Error("m3")
	})

	logLines := splitLines(logContent)
	if len(logLines) != numErrors {
		t.Errorf("Wrong number of log lines on file: expected '%d', received '%d'", numErrors, len(logLines))
		return
	}

	screenLines := splitLines(screenContent)
	if len(screenLines) != numErrors {
		t.Errorf("Wrong number of log lines on screen: expected '%d', received '%d'", numErrors, len(screenLines))
		return
	}

	for i := 0; i < numErrors; i++ {
		file := logLines[i]
		screen := screenLines[i]

		if !strings.Contains(file, "stack") {
			t.Errorf("Case %d, log line on file should record a stack", i)
		}

		if strings.Contains(screen, "stack") {
			t.Errorf("Case %d, log line on screen should not record a stack", i)
		}
	}
}

func TestPrintErrorPrintsOnce(t *testing.T) {
	logContent, screenContent := collectLog(t, func() {
		log.SetStdoutLevel(log.LevelOff)

		log.PrintError(errors.New("some error"), "friendly message")
	})

	logLines := splitLines(logContent)

	if len(logLines) > 1 {
		t.Errorf("Log file should have only one line")
	}

	if strings.Contains(logContent, "friendly message") {
		t.Errorf("File log should not register the user-friendly message")
	}

	if !strings.Contains(logContent, "some error") {
		t.Errorf("File log should register the raw error message")
	}

	screenLines := splitLines(screenContent)

	if len(screenLines) > 1 {
		t.Errorf("Stdout log should have only one line")
	}

	if !strings.Contains(screenContent, "friendly message") {
		t.Errorf("Stdout log should register the user-friendly message")
	}

	if strings.Contains(screenContent, "some error") {
		t.Errorf("Stdout log should not register the raw error message")
	}
}

func TestPrintErrorStackFrame(t *testing.T) {
	logContent, _ := collectLog(t, func() {
		log.PrintError(auxStack(false, false), "e1")
		log.PrintError(auxStack(false, true), "e2")
		log.PrintError(auxStack(true, false), "e3")
		log.PrintError(auxStack(true, true), "e4")
	})

	logLines := splitLines(logContent)

	for i, line := range logLines {
		if strings.Contains(line, "log.PrintError") {
			t.Errorf("Case %d, wrong stack when using PrintError. Line:\n%s", i, line)
		}
	}
}
