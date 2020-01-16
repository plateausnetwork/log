package log_test

import (
	"testing"

	"github.com/rhizomplatform/log"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    log.Level
		expected string
	}{
		{level: log.LevelDebug, expected: "debug"},
		{level: log.LevelInfo, expected: "info"},
		{level: log.LevelWarn, expected: "warning"},
		{level: log.LevelError, expected: "error"},
		{level: log.LevelOff, expected: "off"},
	}

	for i, test := range tests {
		actual := test.level.String()

		if actual != test.expected {
			t.Errorf("Case %d, expected '%s', received '%s'", i, test.expected, actual)
		}
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		str      string
		expected log.Level
		hasError bool
	}{
		{str: "fatal", hasError: true},
		{str: "panic", hasError: true},
		{str: "debug", expected: log.LevelDebug},
		{str: "info", expected: log.LevelInfo},
		{str: "warn", expected: log.LevelWarn},
		{str: "warning", expected: log.LevelWarn},
		{str: "error", expected: log.LevelError},
		{str: "off", expected: log.LevelOff},
		{str: "", hasError: true},
		{str: "foo", hasError: true},
	}

	for i, test := range tests {
		actual, err := log.ParseLevel(test.str)

		if err != nil && !test.hasError {
			t.Errorf("Case %d, error parsing level: %v", i, err)
		} else if err == nil && test.hasError {
			t.Errorf("Case %d, parse should return error", i)
		}

		if err == nil && actual != test.expected {
			t.Errorf("Case %d, expected '%s', received '%s'", i, test.expected.String(), actual.String())
		}
	}
}
