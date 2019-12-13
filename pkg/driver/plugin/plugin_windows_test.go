// +build windows

package plugin_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/graphql-editor/stucco/pkg/driver/plugin"
	"github.com/stretchr/testify/mock"
)

func TestLoadDriverPluginsCallsConfig(t *testing.T) {
	execMock := &execCommandContextMock{}
	plugin.ExecCommandContext = execMock.CommandContext
	oldPath := os.Getenv("PATH")
	// ignores bad paths, and falls back to cwd for empty path
	os.Setenv("PATH", string(os.PathListSeparator)+"/bad/path")

	// fake executable
	f, _ := os.Create("stucco-fake-plugin.exe")
	f.Close()

	// fake script
	f, _ = os.Create("stucco-fake-plugin.cmd")
	f.Close()

	// fake bad plugin does not cause an error
	f, _ = os.Create("stucco-fake-bad-plugin.exe")
	f.Close()

	// non executables are ignored
	f, _ = os.Create("stucco-not-plugin")
	f.Close()

	// directories are ignored
	os.Mkdir("stucco-dir", 0777)
	defer func() {
		plugin.ExecCommandContext = exec.CommandContext
		os.Setenv("PATH", oldPath)
		os.Remove("stucco-fake-plugin.exe")
		os.Remove("stucco-fake-plugin.cmd")
		os.Remove("stucco-fake-bad-plugin.exe")
		os.Remove("stucco-not-plugin")
		os.Remove("stucco-dir.exe")
	}()
	execMock.On(
		"CommandContext",
		mock.Anything,
		"stucco-fake-plugin.exe",
		"config",
	).Return(fakeExecCommandContext)
	execMock.On(
		"CommandContext",
		mock.Anything,
		"stucco-fake-plugin.cmd",
		"config",
	).Return(fakeExecCommandContext)
	execMock.On(
		"CommandContext",
		mock.Anything,
		"stucco-fake-bad-plugin.exe",
		"config",
	).Return(fakeBadExecCommandContext)
	cleanup := plugin.LoadDriverPlugins(plugin.Config{})
	cleanup()
	execMock.AssertCalled(
		t,
		"CommandContext",
		mock.Anything,
		"stucco-fake-plugin.exe",
		"config",
	)
	execMock.AssertCalled(
		t,
		"CommandContext",
		mock.Anything,
		"stucco-fake-plugin.cmd",
		"config",
	)
	execMock.AssertCalled(
		t,
		"CommandContext",
		mock.Anything,
		"stucco-fake-bad-plugin.exe",
		"config",
	)
}
