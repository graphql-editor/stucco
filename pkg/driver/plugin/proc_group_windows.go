package plugin

import (
	"context"
	"os/exec"
	"strconv"
	"time"
)

func createProcGroup(cmd *exec.Cmd) {}

func killTree(cmd *exec.Cmd) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return exec.
		CommandContext(ctx, "taskkill", "/pid", strconv.FormatInt(int64(cmd.Process.Pid), 10), "/T", "/F").
		Run()
}
