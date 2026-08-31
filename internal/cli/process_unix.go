//go:build !windows

package cli

import (
	"os"
	"os/exec"
	"syscall"
)

// setupProcessGroup (GIN-029): the child gets its own process group so the
// signal can reach `go run` AND the compiled binary (the grandchild) — `go
// run` does not reliably forward signals to the app.
func setupProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// signalChild sends the signal to the whole child process group.
func signalChild(cmd *exec.Cmd, sig os.Signal) error {
	if s, ok := sig.(syscall.Signal); ok && cmd.Process != nil {
		return syscall.Kill(-cmd.Process.Pid, s)
	}
	if cmd.Process != nil {
		return cmd.Process.Signal(sig)
	}
	return nil
}
