//go:build linux

package supervise

import (
	"errors"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func configureDaemonReaping() error {
	err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0)
	if err == nil {
		return nil
	}
	if errors.Is(err, unix.EINVAL) || errors.Is(err, unix.EPERM) {
		return nil
	}
	return err
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
	// The wrapped direct child may exit before descendants in the same process
	// group. Reap the remaining process group before _exec exits so npm/node
	// trees do not survive long enough to be orphaned to PID 1.
	if err := cancelDaemonCommand(cmd); err != nil {
		return err
	}
	if err := reapDaemonChildren(); err != nil {
		return err
	}
	return waitErr
}

func reapDaemonChildren() error {
	hardDeadline := time.Now().Add(10 * time.Second)
	idleDeadline := time.Now().Add(2 * time.Second)
	for {
		if time.Now().After(hardDeadline) {
			return nil
		}
		var status unix.WaitStatus
		pid, err := unix.Wait4(-1, &status, unix.WNOHANG, nil)
		if err == nil {
			if pid > 0 {
				idleDeadline = time.Now().Add(2 * time.Second)
				continue
			}
			if time.Now().After(idleDeadline) {
				return nil
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}
		if errors.Is(err, unix.ECHILD) {
			return nil
		}
		if errors.Is(err, unix.EINTR) {
			continue
		}
		return err
	}
}
