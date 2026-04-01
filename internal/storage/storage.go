package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const ConfigDirEnv = "TAXSEND_CONFIG_DIR"

var validNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

func RootDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv(ConfigDirEnv)); override != "" {
		return override, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "taxsend"), nil
}

func ReceiversDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "receivers"), nil
}

func ProfilesDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profiles"), nil
}

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !validNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid name %q: use letters, numbers, dot, underscore, or hyphen", name)
	}
	return name, nil
}

func ReceiverMetadataPath(name string) (string, error) {
	name, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	dir, err := ReceiversDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func ReceiverKeystorePath(name string) (string, error) {
	name, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	dir, err := ReceiversDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".age"), nil
}

func ProfilePath(name string) (string, error) {
	name, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	dir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func ListJSONNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}
