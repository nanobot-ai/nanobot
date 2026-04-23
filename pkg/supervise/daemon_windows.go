//go:build windows

package supervise

import "os/exec"

func configureDaemonCommand(*exec.Cmd) {
}

func cancelDaemonCommand(cmd *exec.Cmd) error {
	if cmd.Process != nil {
		return cmd.Process.Kill()
	}
	return nil
}
