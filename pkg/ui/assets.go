package ui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/log"
)

//go:embed assets
var Assets embed.FS

func copyFileFromAssets(dest string, path string) error {
	data, err := Assets.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(dest, data, 0644)
}

func copyFiles(dest string, hash string) error {
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm not found: %w", err)
	}

	target := filepath.Join(dest, hash)
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("failed to remove target: %w", err)
	}

	target += ".tmp"
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("failed to remove temp target: %w", err)
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}

	prefix := "assets/" + hash
	err := fs.WalkDir(Assets, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		target := filepath.Join(target, strings.TrimPrefix(path, prefix))
		return copyFileFromAssets(target, path)
	})
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	installCmd := exec.Command("npm", "install", "--omit=dev", "--ignore-scripts", "--no-audit", "--no-fund")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Dir = target
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	return os.Rename(target, strings.TrimSuffix(target, ".tmp"))
}

func StartUI(ctx context.Context, mcpAddress string) (http.Handler, error) {
	path, err := StageUI()
	if err != nil {
		return nil, fmt.Errorf("failed to stage UI: %w", err)
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	port := l.Addr().(*net.TCPAddr).Port

	if err := l.Close(); err != nil {
		return nil, fmt.Errorf("failed to close listener: %w", err)
	}

	cmd := exec.CommandContext(ctx, "node", "./node_modules/.bin/react-router-serve", "./build/server/index.js")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path
	cmd.Env = append(os.Environ(), fmt.Sprintf("NANOBOT_URL=%s", mcpAddress),
		"NODE_ENV=production",
		"HOST=127.0.0.1",
		"PORT="+fmt.Sprintf("%d", port))
	context.AfterFunc(ctx, func() {
		_ = cmd.Process.Kill()
	})
	go func() {
		if err := cmd.Run(); err != nil {
			log.Fatalf(context.Background(), "Failed to run UI: %v", err)
		}
	}()

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", port),
	})
	return proxy, nil
}

func StageUI() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine cache directory: %w", err)
	}

	dir = filepath.Join(dir, "nanobot/ui")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	files, err := Assets.ReadDir("assets")
	if err != nil {
		return "", fmt.Errorf("failed to read assets: %w", err)
	}

	for _, file := range files {
		ret := filepath.Join(dir, file.Name())
		if _, err := os.Stat(filepath.Join(dir, file.Name())); err == nil {
			return ret, nil
		}
		if err := copyFiles(dir, file.Name()); err != nil {
			return "", fmt.Errorf("failed to copy dir: %w", err)
		}
		return ret, nil
	}

	return "", fmt.Errorf("no UI assets found, run `go generate` to build")
}
