//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// setupProcessGroup is a no-op on Windows (no process groups).
func setupProcessGroup(cmd *exec.Cmd) {}

// signalChild signals only the direct child on Windows.
func signalChild(cmd *exec.Cmd, sig os.Signal) error {
	if cmd.Process != nil {
		return cmd.Process.Signal(sig)
	}
	return nil
}
