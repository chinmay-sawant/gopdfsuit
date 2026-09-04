//go:build unix

package redact

import (
	"context"
	"os/exec"
	"syscall"
)

// ocrCommand builds a context-bound command that kills the whole process
// group on timeout. Plain CommandContext only kills the direct child, but a
// shell wrapper (or a helper that forks) would otherwise keep pipe FDs open
// and CombinedOutput would block until the grandchildren exit.
func ocrCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		_ = syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	return cmd
}
