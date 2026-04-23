//go:build windows

package supervise

import "os/exec"

func configureDaemonReaping() error {
	return nil
}

func configureDaemonCommand(*exec.Cmd) {
}

func cancelDaemonCommand(cmd *exec.Cmd) error {
	if cmd.Process != nil {
		return cmd.Process.Kill()
	}
	return nil
}

func afterDaemonCommandExit(_ *exec.Cmd, waitErr error) error {
	return waitErr
}
