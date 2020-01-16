package log_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rhizomplatform/fs"
	"github.com/rhizomplatform/log"
)

func TestLoggerSingleton(t *testing.T) {
	baseFolder, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("error creating temp directory:", err)
	}
	defer fs.RemoveAll(baseFolder)

	tests := []struct {
		folder      string
		shouldExist bool
	}{
		{folder: uuid.New().String(), shouldExist: false},
		{folder: uuid.New().String(), shouldExist: false},
		{folder: uuid.New().String(), shouldExist: false},
		{folder: uuid.New().String(), shouldExist: false},
		{folder: uuid.New().String(), shouldExist: false},
	}

	testSetupHelper := func(index int) {
		tests[index].shouldExist = true
		log.Setup(fs.Path(baseFolder).Join(tests[index].folder), "mysufix", 2, 1)
		defer log.TearDown()
		defer fs.Path(baseFolder).Join(tests[index].folder).RemoveAll()

		for i, test := range tests {
			fullPath := fs.Path(baseFolder).Join(test.folder)
			log.Setup(fullPath, "mysufix", 2, 1)
			exists := fullPath.DirExists()

			if !exists && test.shouldExist {
				t.Errorf("[%d] Case %d, directory '%s' should exist", index, i, fullPath)
			} else if exists && !test.shouldExist {
				t.Errorf("[%d] Case %d, directory '%s' should not exist", index, i, fullPath)
			}
		}

		tests[index].shouldExist = false
	}

	testSetupHelper(0)
	testSetupHelper(2)
	testSetupHelper(4)
}

func TestOutputRedirect(t *testing.T) {
	baseFolder, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("error creating temp directory:", err)
	}
	defer fs.RemoveAll(baseFolder)

	log.Setup(fs.Path(baseFolder), "mysufix", 2, 1)
	defer log.TearDown()

	var buffer bytes.Buffer
	screen := bufio.NewWriter(&buffer)

	log.SetStdoutLevel(log.LevelDebug)
	log.RedirectStdout(screen)
	defer log.RestoreStdout()

	tests := []struct {
		writerFn func(string)
		content  string
	}{
		{writerFn: log.Debug, content: "log-debug"},
		{writerFn: log.Info, content: "info-debug"},
		{writerFn: log.Warn, content: "warn-debug"},
	}

	for _, test := range tests {
		test.writerFn(test.content)
	}

	screen.Flush()
	screenContent := buffer.String()

	for i, test := range tests {
		if !strings.Contains(screenContent, test.content) {
			t.Errorf("Case %d, log line '%s' not found", i, test.content)
		}
	}
}

func TestGetStdoutLevel(t *testing.T) {
	baseFolder, err := ioutil.TempDir("", "")
	if err != nil {
		t.Error("error creating temp directory:", err)
	}
	defer fs.RemoveAll(baseFolder)

	log.Setup(fs.Path(baseFolder), "mysufix", 2, 1)
	defer log.TearDown()

	tests := []log.Level{
		log.LevelOff,
		log.LevelInfo,
		log.LevelDebug,
		log.LevelOff,
		log.LevelWarn,
		log.LevelError,
	}

	for i, test := range tests {
		log.SetStdoutLevel(test)
		level := log.GetStdoutLevel()

		if level != test {
			t.Errorf("Case %d, expected level '%s', but received '%s'", i, test, level)
		}
	}
}
