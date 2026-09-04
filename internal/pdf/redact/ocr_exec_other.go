//go:build !unix

package redact

import (
	"context"
	"os/exec"
)

// ocrCommand is the portable fallback for platforms without process groups
// (Windows, js/wasm): a plain context-bound command without group kill.
func ocrCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
