package fsutil

import (
	"errors"
	"fmt"
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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
