package supervise

import (
	"context"
	"os"
	"os/exec"
	"time"
)

func Daemon() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := configureDaemonReaping(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, os.Args[2], os.Args[3:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	configureDaemonCommand(cmd)

	cmd.Cancel = func() error {
		return cancelDaemonCommand(cmd)
	}

	processIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		_ = processIn.Close()
		return err
	}

	go func() {
		defer processIn.Close()

		var buf [4096]byte
		for {
			n, err := os.Stdin.Read(buf[:])
			if err != nil {
				break
			}
			if n > 0 {
				if _, err := processIn.Write(buf[:n]); err != nil {
					break
				}
			}
		}
		// Give the wrapped command a short grace period to exit on stdin EOF before
		// canceling and escalating to process-group termination.
		time.Sleep(5 * time.Second)
		cancel()
	}()

	err = cmd.Wait()
	return afterDaemonCommandExit(cmd, err)
}
