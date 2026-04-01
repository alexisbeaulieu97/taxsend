package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func EnsureNotExists(path string, force bool) error {
	_, err := os.Stat(path)
	if err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func CreateAtomic(path string, force bool) (*os.File, string, error) {
	if err := EnsureNotExists(path, force); err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".taxsend-*")
	if err != nil {
		return nil, "", err
	}
	return tmp, tmp.Name(), nil
}

func CommitAtomic(tmpPath, finalPath string) error {
	return os.Rename(tmpPath, finalPath)
}

func PrepareOutputPath(path string, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	_, err := os.Stat(path)
	if err == nil {
		if !force {
			return fmt.Errorf("%s already exists (use --force to overwrite)", path)
		}
		return os.Remove(path)
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func CommitAtomicForce(tmpPath, finalPath string, force bool) error {
	if err := PrepareOutputPath(finalPath, force); err != nil {
		return err
	}
	return os.Rename(tmpPath, finalPath)
}

func WriteFileAtomic(path string, data []byte, perm fs.FileMode, force bool) error {
	tmp, tmpPath, err := CreateAtomic(path, force)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return CommitAtomicForce(tmpPath, path, force)
}

func EnsurePrivateDir(path string) error {
	return os.MkdirAll(path, 0o700)
}

func SafeJoin(baseDir, tarPath string) (string, error) {
	if filepath.IsAbs(tarPath) {
		return "", fmt.Errorf("archive contains absolute path: %s", tarPath)
	}
	clean := filepath.Clean(tarPath)
	if clean == "." || clean == "" {
		return "", fmt.Errorf("invalid archive path: %s", tarPath)
	}
	target := filepath.Join(baseDir, clean)
	rel, err := filepath.Rel(baseDir, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || rel == "." || rel[:min(3, len(rel))] == ".."+string(filepath.Separator) {
		return "", fmt.Errorf("unsafe archive path: %s", tarPath)
	}
	return target, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
