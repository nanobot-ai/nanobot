package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func hash() (string, error) {
	r, err := os.OpenRoot("./ui")
	if err != nil {
		return "", err
	}
	defer r.Close()

	digest := sha256.New()
	err = fs.WalkDir(r.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		ok := strings.HasPrefix(path, "package.json") ||
			strings.HasPrefix(path, "package-lock.json") ||
			strings.HasPrefix(path, "app/")
		if !ok {
			return nil
		}
		f, err := r.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer f.Close()
		_, err = io.Copy(digest, f)
		return err
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func build() error {
	cmd := exec.Command("npm", "install", "--ignore-scripts", "--no-audit", "--no-fund")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "./ui"
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	cmd = exec.Command("npm", "run", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "./ui"
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}

	return nil
}

func copyFile(src fs.FS, dest *os.Root, path string) error {
	srcFile, err := src.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", path, err)
	}
	defer srcFile.Close()

	if err := dest.Mkdir(filepath.Dir(path), 0755); err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(path), err)
	}

	destFile, err := dest.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination file %s: %w", path, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file %s: %w", path, err)
	}

	return nil
}

func copyDir(hash string) error {
	err := os.RemoveAll("./pkg/ui/assets")
	if err != nil {
		return fmt.Errorf("failed to remove assets: %w", err)
	}

	src, err := os.OpenRoot("./ui")
	if err != nil {
		return fmt.Errorf("failed to open root: %w", err)
	}
	defer src.Close()

	if err := os.MkdirAll("./pkg/ui/assets/"+hash+"/build/client/assets", 0755); err != nil {
		return fmt.Errorf("failed to create assets: %w", err)
	}

	if err := os.MkdirAll("./pkg/ui/assets/"+hash+"/build/server", 0755); err != nil {
		return fmt.Errorf("failed to create assets: %w", err)
	}

	dest, err := os.OpenRoot("./pkg/ui/assets/" + hash)
	if err != nil {
		return fmt.Errorf("failed to open dest: %w", err)
	}
	defer dest.Close()

	if err := copyFile(src.FS(), dest, "package.json"); err != nil {
		return fmt.Errorf("failed to copy package.json: %w", err)
	}

	if err := copyFile(src.FS(), dest, "package-lock.json"); err != nil {
		return fmt.Errorf("failed to copy package.json: %w", err)
	}

	return fs.WalkDir(src.FS(), "build", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		return copyFile(src.FS(), dest, path)
	})
}

func main() {
	hash, err := hash()
	if err != nil {
		panic(err)
	}

	_, err = os.Stat("./pkg/ui/assets/" + hash)
	if errors.Is(err, fs.ErrNotExist) {
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("Assets for ui up to date (" + hash + ").")
		return
	}

	fmt.Println("Generating assets for ui (" + hash + ") ...")

	if err := build(); err != nil {
		panic(err)
	}

	if err := copyDir(hash + ".tmp"); err != nil {
		panic(err)
	}

	if err := os.Rename("./pkg/ui/assets/"+hash+".tmp", "./pkg/ui/assets/"+hash); err != nil {
		panic(fmt.Errorf("failed to rename assets: %w", err))
	}

	fmt.Println(hash)
}
