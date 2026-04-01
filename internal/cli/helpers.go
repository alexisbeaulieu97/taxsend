package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"taxsend/internal/profile"
)

func defaultAttachmentBase(ts time.Time) string {
	return "attachment-" + ts.Format("20060102-150405")
}

func defaultUnsealDir(ts time.Time) string {
	return "unsealed-" + ts.Format("20060102-150405")
}

func defaultPublicProfilePath(name string) string {
	return name + ".public.json"
}

func normalizeBasename(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("basename is required")
	}
	if filepath.Base(value) != value {
		return "", fmt.Errorf("basename must not include path separators")
	}
	ext := filepath.Ext(value)
	if ext != "" {
		value = strings.TrimSuffix(value, ext)
	}
	value = strings.TrimSpace(value)
	if value == "" || value == "." || value == ".." {
		return "", fmt.Errorf("invalid basename %q", value)
	}
	return value, nil
}

func resolveReceiverName(name string) (string, error) {
	if strings.TrimSpace(name) != "" {
		return name, nil
	}
	names, err := profile.ListReceiverNames()
	if err != nil {
		return "", err
	}
	switch len(names) {
	case 0:
		return "", fmt.Errorf("no local receiver profiles found")
	case 1:
		return names[0], nil
	default:
		return "", fmt.Errorf("multiple receiver profiles found; use --name")
	}
}
