package log_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/rhizomplatform/log"
)

func TestWithEntry(t *testing.T) {
	tests := make([]string, 0)

	logContent, _ := collectLog(t, func() {
		log.With(log.F{}).With(log.F{"msg-x": 1}).Debug("debug-msg")
		tests = append(tests, fmt.Sprintf("\"msg\":\"%s\",\"msg-x\":%v", "debug-msg", 1))

		log.With(log.F{}).With(log.F{"msg-y": 2}).Info("info-msg")
		tests = append(tests, fmt.Sprintf("\"msg\":\"%s\",\"msg-y\":%v", "info-msg", 2))

		log.With(log.F{}).With(log.F{"msg-z": "3"}).Warn("warn-msg")
		tests = append(tests, fmt.Sprintf("\"msg\":\"%s\",\"msg-z\":\"%v\"", "warn-msg", 3))
	})

	for i, test := range tests {
		if !strings.Contains(logContent, test) {
			t.Errorf("Case %d, log line '%s' not found", i, test)
		}
	}
}

func TestEntryError(t *testing.T) {
	tests := []struct {
		msg        string
		errMsg     string
		field      string
		errorFirst bool
	}{
		{msg: "error1"},
		{msg: "error2", errMsg: "emsg2"},
		{msg: "error3", field: "f3"},
		{msg: "error4", field: "f4", errMsg: "emsg4"},
		{msg: "error5", field: "f5", errMsg: "emsg5", errorFirst: true},
	}

	logContent, _ := collectLog(t, func() {
		for _, test := range tests {
			switch {
			case test.errMsg != "" && test.field != "" && test.errorFirst:
				log.WithError(errors.New(test.errMsg)).With(log.F{"field": test.field}).Error(test.msg)
			case test.errMsg != "" && test.field != "":
				log.With(log.F{"field": test.field}).WithError(errors.New(test.errMsg)).Error(test.msg)
			case test.errMsg != "":
				log.WithError(errors.New(test.errMsg)).Error(test.msg)
			case test.field != "":
				log.With(log.F{"field": test.field}).Error(errors.New(test.msg))
			default:
				log.Error(errors.New(test.msg))
			}
		}
	})

	logLines := splitLines(logContent)

	if len(logLines) != len(tests) {
		t.Errorf("Wrong number of log lines: expected '%d', received '%d'", len(tests), len(logLines))
		t.Log(logLines)
		return
	}

	for i, test := range tests {
		line := logLines[i]

		if test.errMsg != "" && !strings.Contains(line, fmt.Sprintf("\"error\":\"%s\"", test.errMsg)) {
			t.Errorf("Case %d, inner error message not found, expected '%s'", i, test.errMsg)
		} else if test.errMsg == "" && strings.Contains(line, "\"error\":") {
			t.Errorf("Case %d, custom error should not be present", i)
		}

		if test.field != "" && !strings.Contains(line, fmt.Sprintf("\"field\":\"%s\"", test.field)) {
			t.Errorf("Case %d, custom field not found, expected '%s'", i, test.field)
		} else if test.field == "" && strings.Contains(line, "\"field\":") {
			t.Errorf("Case %d, custom field should not be present", i)
		}

		if !strings.Contains(line, fmt.Sprintf("\"msg\":\"%s\"", test.msg)) {
			t.Errorf("Case %d, error message '%s' not found", i, test.msg)
		}

		if !strings.Contains(line, "\"stack\"") {
			t.Errorf("Case %d, stacktrace should not be empty", i)
		}
	}
}

// exists only to test Stringer support on error collection
type someStringer struct{}

func (s someStringer) String() string {
	return "something"
}

func TestEntryErrorRewrite(t *testing.T) {
	logContent, _ := collectLog(t, func() {
		log.WithError(errors.New("foo")).With(log.F{"error": "a string"}).Error("e1")
		log.WithError(errors.New("foo")).With(log.F{"error": someStringer{}}).Error("e1")
	})

	if !strings.Contains(logContent, "\"error\":\"a string\"") {
		t.Errorf("Error field should gracefully handle strings")
	}

	if !strings.Contains(logContent, "\"error\":\"something\"") {
		t.Errorf("Error field should gracefully handle Stringer interface")
	}
}
