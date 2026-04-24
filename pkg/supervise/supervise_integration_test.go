//go:build integration && !windows

package supervise_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestExecSupervisorCleansUpGrandchildrenOnStdinClose(t *testing.T) {
	repoRoot := superviseRepoRoot(t)
	binPath := buildNanobotBinary(t, repoRoot)
	helperPath, pidFile := writeGrandchildHelper(t)

	cmd, stdin, output := startExecSupervisor(t, binPath, helperPath, supervisorStartOptions{
		withStdin:           true,
		supervisorOwnPGroup: true,
	})
	go func() {
		time.Sleep(1 * time.Second)
		_ = stdin.Close()
	}()
	waitForExit(t, cmd, output)

	assertGrandchildExited(t, pidFile, "stdin close in shared-group regime")
}

func TestExecSupervisorCleansUpGrandchildrenOnStdinCloseWithSplitChildGroup(t *testing.T) {
	repoRoot := superviseRepoRoot(t)
	binPath := buildNanobotBinary(t, repoRoot)
	helperPath, pidFile := writeGrandchildHelper(t)

	cmd, stdin, output := startExecSupervisor(t, binPath, helperPath, supervisorStartOptions{
		withStdin:           true,
		supervisorOwnPGroup: false,
	})
	go func() {
		time.Sleep(1 * time.Second)
		_ = stdin.Close()
	}()
	waitForExit(t, cmd, output)

	assertGrandchildExited(t, pidFile, "stdin close in split-child-group regime")
}

func TestExecSupervisorCleansUpGrandchildrenWhenDirectChildExitsOnEOF(t *testing.T) {
	repoRoot := superviseRepoRoot(t)
	binPath := buildNanobotBinary(t, repoRoot)
	helperPath, pidFile := writeEarlyExitGrandchildHelper(t)

	cmd, stdin, output := startExecSupervisor(t, binPath, helperPath, supervisorStartOptions{
		withStdin:           true,
		supervisorOwnPGroup: true,
	})

	stdinWriter, ok := stdin.(io.WriteCloser)
	if !ok {
		t.Fatalf("stdin pipe does not support writes")
	}
	if _, err := io.WriteString(stdinWriter, "hello\n"); err != nil {
		t.Fatalf("failed to write to supervisor stdin: %v", err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		_ = stdin.Close()
	}()

	waitForExit(t, cmd, output)
	assertGrandchildExited(t, pidFile, "direct child exit after stdin EOF")
}

func TestExecSupervisorCleansUpGrandchildrenWhenDirectChildExitsImmediately(t *testing.T) {
	repoRoot := superviseRepoRoot(t)
	binPath := buildNanobotBinary(t, repoRoot)
	helperPath, pidFile := writeImmediateExitGrandchildHelper(t)

	cmd, _, output := startExecSupervisor(t, binPath, helperPath, supervisorStartOptions{
		withStdin:           false,
		supervisorOwnPGroup: true,
	})

	waitForExit(t, cmd, output)
	assertGrandchildExited(t, pidFile, "direct child immediate exit")
}

type supervisorStartOptions struct {
	withStdin           bool
	supervisorOwnPGroup bool
}

func startExecSupervisor(t *testing.T, binPath, helperPath string, opts supervisorStartOptions) (*exec.Cmd, io.Closer, *strings.Builder) {
	t.Helper()

	cmd := exec.Command(binPath, "_exec", helperPath)
	if opts.supervisorOwnPGroup {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	var stdin io.Closer
	if opts.withStdin {
		var err error
		stdin, err = cmd.StdinPipe()
		if err != nil {
			t.Fatalf("failed to create stdin pipe: %v", err)
		}
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start nanobot _exec: %v", err)
	}
	return cmd, stdin, &output
}

func waitForExit(t *testing.T, cmd *exec.Cmd, output *strings.Builder) {
	t.Helper()

	err := cmd.Wait()
	if err == nil {
		t.Logf("nanobot _exec output:\n%s", strings.TrimSpace(output.String()))
	} else {
		t.Logf("nanobot _exec exited with %v\n%s", err, strings.TrimSpace(output.String()))
	}
}

func writeGrandchildHelper(t *testing.T) (helperPath, pidFile string) {
	t.Helper()

	workDir := t.TempDir()
	pidFile = filepath.Join(workDir, "grandchild.pid")
	helperPath = filepath.Join(workDir, "spawn-grandchild.sh")
	writeExecutableFile(t, helperPath, fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
sleep 1000 </dev/null >/dev/null 2>&1 &
grandchild=$!
printf '%%s\n' "${grandchild}" > %q
wait
`, pidFile))
	return helperPath, pidFile
}

func writeEarlyExitGrandchildHelper(t *testing.T) (helperPath, pidFile string) {
	t.Helper()

	workDir := t.TempDir()
	pidFile = filepath.Join(workDir, "grandchild.pid")
	helperPath = filepath.Join(workDir, "spawn-grandchild-early-exit.sh")
	writeExecutableFile(t, helperPath, fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
sleep 1000 </dev/null >/dev/null 2>&1 &
grandchild=$!
printf '%%s\n' "${grandchild}" > %q
cat >/dev/null
exit 0
`, pidFile))
	return helperPath, pidFile
}

func writeImmediateExitGrandchildHelper(t *testing.T) (helperPath, pidFile string) {
	t.Helper()

	workDir := t.TempDir()
	pidFile = filepath.Join(workDir, "grandchild.pid")
	helperPath = filepath.Join(workDir, "spawn-grandchild-immediate-exit.sh")
	writeExecutableFile(t, helperPath, fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
sleep 1000 </dev/null >/dev/null 2>&1 &
grandchild=$!
printf '%%s\n' "${grandchild}" > %q
exit 0
`, pidFile))
	return helperPath, pidFile
}

func waitForPIDFile(t *testing.T, pidFile string) int {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for {
		pidBytes, err := os.ReadFile(pidFile)
		if err == nil && strings.TrimSpace(string(pidBytes)) != "" {
			pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
			if err != nil {
				t.Fatalf("failed to parse pid from %s: %v", pidFile, err)
			}
			return pid
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("failed to read pid file %s: %v", pidFile, err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		t.Fatalf("failed to parse pid from %s: %v", pidFile, err)
	}
	return pid
}

func assertGrandchildExited(t *testing.T, pidFile string, reason string) {
	t.Helper()

	pid := waitForPIDFile(t, pidFile)
	defer killIfAlive(t, pid)

	time.Sleep(1 * time.Second)

	if processExists(pid) {
		psOut, _ := exec.Command("ps", "-o", "pid=,ppid=,pgid=,command=", "-p", strconv.Itoa(pid)).CombinedOutput()
		t.Fatalf("grandchild process %d is still alive after %s\nprocess:\n%s", pid, reason, strings.TrimSpace(string(psOut)))
	}
}

func superviseRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func buildNanobotBinary(t *testing.T, repoRoot string) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "nanobot-test-bin")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build nanobot binary: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	return binPath
}

func writeExecutableFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}

func killIfAlive(t *testing.T, pid int) {
	t.Helper()
	if !processExists(pid) {
		return
	}
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		t.Logf("failed to kill leaked process %d: %v", pid, err)
	}
}
