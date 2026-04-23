//go:build !windows

package supervise

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureDaemonCommand(cmd *exec.Cmd) {
	wrapperPGID, err := syscall.Getpgid(0)
	if err != nil || wrapperPGID != os.Getpid() {
		// There are two regimes here:
		//   1. If _exec is already the leader of its own process group, keep the wrapped
		//      command in that same group so an external kill of the wrapper group
		//      naturally reaps the whole subtree.
		//   2. Otherwise, put the wrapped command in its own group so Cancel can still
		//      signal the entire child subtree without depending on the parent's group.
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
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
