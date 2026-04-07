//go:build ignore

package testdata

import (
	"context"
	"os/exec"
)

func BadShell() {
	exec.Command("sh", "-c", "rm -rf /")
	exec.CommandContext(context.Background(), "sh", "-c", "echo hi")
}
