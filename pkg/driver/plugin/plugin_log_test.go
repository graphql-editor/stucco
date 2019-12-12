package plugin_test

import (
	"flag"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/stretchr/testify/assert"
	"k8s.io/klog"
)

func TestHCLogToKLogRedirect(t *testing.T) {
	flagSet := flag.NewFlagSet("klogflags", flag.ContinueOnError)
	klog.InitFlags(flagSet)
	flagSet.Parse([]string{"-v=5"})
	oldStderr := os.Stderr
	defer func() {
		os.Stderr = oldStderr
	}()
	pr, pw, _ := os.Pipe()
	os.Stderr = pw
	data := []struct {
		level    string
		inMsg    string
		inArgs   []interface{}
		expected string
	}{
		{
			level:    "Trace",
			inMsg:    "trace message\ntrace message line 2",
			inArgs:   []interface{}{"arg1", "value1"},
			expected: "^E.*] logger.name: trace message$\ntrace message line 2: arg1=value1",
		},
		{
			level:    "Info",
			inMsg:    "info message\ninfo message line 2",
			inArgs:   []interface{}{"arg1", "value1"},
			expected: "^I.*] logger.name: info message$\ninfo message line 2: arg1=value1",
		},
		{
			level:    "Debug",
			inMsg:    "debug message\ndebug message line 2",
			inArgs:   []interface{}{"arg1", "value1"},
			expected: "^I.*] logger.name: debug message$\ndebug message line 2: arg1=value1",
		},
		{
			level:    "Warn",
			inMsg:    "warn message\nwarn message line 2",
			inArgs:   []interface{}{"arg1", "value1"},
			expected: "^W.*] logger.name: warn message$\nwarn message line 2: arg1=value1",
		},
		{
			level:    "Error",
			inMsg:    "error message\nerror message line 2",
			inArgs:   []interface{}{"arg1", "value1"},
			expected: "^E.*] logger.name: error message$\nerror message line 2: arg1=value1",
		},
	}
	logger := plugin.NewLogger("logger").Named("name")
	t.Run("Levels", func(t *testing.T) {
		for i := range data {
			tt := data[i]
			t.Run(tt.level, func(t *testing.T) {
				t.Parallel()
				var f func(string, ...interface{})
				switch tt.level {
				case "Trace":
					f = logger.Trace
				case "Info":
					f = logger.Info
				case "Debug":
					f = logger.Debug
				case "Warn":
					f = logger.Warn
				case "Error":
					f = logger.Error
				}
				f(tt.inMsg, tt.inArgs...)
			})
		}
	})
	klog.Flush()
	pw.Close()
	b, _ := ioutil.ReadAll(pr)
	lines := strings.Split(string(b), "\n")
	lines = lines[:len(lines)-1]
	for _, tt := range data {
		assert.Condition(t, func() bool {
			expectedLines := strings.Split(tt.expected, "\n")
			re := regexp.MustCompile(expectedLines[0])
			for i := 0; i < len(lines); i++ {
				if re.Match([]byte(lines[i])) {
					forward := 1
					for j := forward; j < len(expectedLines); j++ {
						if lines[i+j] == expectedLines[j] {
							forward = j + 1
						}
					}
					if forward == len(expectedLines) {
						lines = append(lines[:i], lines[i+forward:]...)
						return true
					}
				}
			}
			return false
		})
	}
	assert.Len(t, lines, 0)
}
