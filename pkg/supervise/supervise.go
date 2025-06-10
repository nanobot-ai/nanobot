package supervise

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/obot-platform/nanobot/pkg/system"
)

func Cmd(ctx context.Context, command string, args ...string) *exec.Cmd {
	args = append([]string{"_exec", command}, args...)
	return exec.CommandContext(ctx, system.Bin(), args...)
}

func Daemon() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[2], os.Args[3:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Cancel = func() error {
		if runtime.GOOS == "windows" {
			return cmd.Process.Kill()
		}
		return cmd.Process.Signal(os.Interrupt)
	}

	processIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
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
		time.Sleep(5 * time.Second)
		cancel()
	}()

	return cmd.Run()
}
