//go:build !windows && !linux

package supervise

import (
	"errors"
	"os/exec"
	"syscall"
)

func configureDaemonReaping() error {
	return nil
}

func configureDaemonCommand(cmd *exec.Cmd) {
	// Always put the wrapped command in its own process group so _exec can kill
	// the child subtree without signaling itself during stdin-EOF cleanup.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func cancelDaemonCommand(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	targetPGID, err := daemonTargetPGID(cmd)
	if err != nil || targetPGID <= 0 {
		return err
	}

	err = syscall.Kill(-targetPGID, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func daemonTargetPGID(cmd *exec.Cmd) (int, error) {
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid {
		return cmd.Process.Pid, nil
	}
	return syscall.Getpgid(0)
}

func afterDaemonCommandExit(cmd *exec.Cmd, waitErr error) error {
	if err := cancelDaemonCommand(cmd); err != nil {
		return err
	}
	return waitErr
}
