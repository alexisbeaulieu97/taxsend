package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"taxsend/internal/crypto"
	"taxsend/internal/fsutil"
	"taxsend/internal/storage"
)

const (
	Version                       = 1
	DefaultPartSizeBytes   int64  = 7 * 1024 * 1024
	DefaultOutputExtension string = ".bin"
)

type SenderProfile struct {
	Version                 int      `json:"version"`
	Name                    string   `json:"name"`
	Recipients              []string `json:"recipients"`
	DefaultMaxPartSizeBytes int64    `json:"default_max_part_size_bytes"`
	DefaultOutputExtension  string   `json:"default_output_extension"`
}

func (p SenderProfile) Normalized() (*SenderProfile, error) {
	name, err := storage.ValidateName(p.Name)
	if err != nil {
		return nil, err
	}
	out := &SenderProfile{
		Version:                 p.Version,
		Name:                    name,
		DefaultMaxPartSizeBytes: p.DefaultMaxPartSizeBytes,
		DefaultOutputExtension:  strings.TrimSpace(p.DefaultOutputExtension),
	}
	if out.Version == 0 {
		out.Version = Version
	}
	if out.Version != Version {
		return nil, fmt.Errorf("unsupported profile version %d", out.Version)
	}
	if out.DefaultMaxPartSizeBytes == 0 {
		out.DefaultMaxPartSizeBytes = DefaultPartSizeBytes
	}
	if out.DefaultMaxPartSizeBytes <= 0 {
		return nil, fmt.Errorf("default_max_part_size_bytes must be greater than zero")
	}
	if out.DefaultOutputExtension == "" {
		out.DefaultOutputExtension = DefaultOutputExtension
	}
	if !strings.HasPrefix(out.DefaultOutputExtension, ".") || strings.Contains(out.DefaultOutputExtension, string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid default_output_extension %q", out.DefaultOutputExtension)
	}

	seen := map[string]struct{}{}
	for _, recipient := range p.Recipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		if _, err := crypto.ParseRecipient(recipient); err != nil {
			return nil, err
		}
		if _, ok := seen[recipient]; ok {
			continue
		}
		seen[recipient] = struct{}{}
		out.Recipients = append(out.Recipients, recipient)
	}
	if len(out.Recipients) == 0 {
		return nil, fmt.Errorf("profile %q has no recipients", out.Name)
	}
	slices.Sort(out.Recipients)
	return out, nil
}

func Default(name string, recipients []string) (*SenderProfile, error) {
	return (SenderProfile{
		Name:                    name,
		Recipients:              recipients,
		DefaultMaxPartSizeBytes: DefaultPartSizeBytes,
		DefaultOutputExtension:  DefaultOutputExtension,
	}).Normalized()
}

func LoadFile(path string) (*SenderProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var profile SenderProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parse profile %s: %w", path, err)
	}
	return profile.Normalized()
}

func SaveFile(path string, p *SenderProfile, force bool) error {
	if p == nil {
		return fmt.Errorf("profile is required")
	}
	normalized, err := p.Normalized()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return fsutil.WriteFileAtomic(path, data, 0o600, force)
}

func SaveSender(p *SenderProfile, force bool) (string, error) {
	normalized, err := p.Normalized()
	if err != nil {
		return "", err
	}
	path, err := storage.ProfilePath(normalized.Name)
	if err != nil {
		return "", err
	}
	return path, SaveFile(path, normalized, force)
}

func SaveReceiverMetadata(p *SenderProfile, force bool) (string, error) {
	normalized, err := p.Normalized()
	if err != nil {
		return "", err
	}
	path, err := storage.ReceiverMetadataPath(normalized.Name)
	if err != nil {
		return "", err
	}
	return path, SaveFile(path, normalized, force)
}

func LoadSender(name string) (*SenderProfile, error) {
	path, err := storage.ProfilePath(name)
	if err != nil {
		return nil, err
	}
	return LoadFile(path)
}

func LoadReceiverMetadata(name string) (*SenderProfile, error) {
	path, err := storage.ReceiverMetadataPath(name)
	if err != nil {
		return nil, err
	}
	return LoadFile(path)
}

func Import(path string, force bool) (*SenderProfile, string, error) {
	p, err := LoadFile(path)
	if err != nil {
		return nil, "", err
	}
	savedPath, err := SaveSender(p, force)
	if err != nil {
		return nil, "", err
	}
	return p, savedPath, nil
}

func ListSenderNames() ([]string, error) {
	dir, err := storage.ProfilesDir()
	if err != nil {
		return nil, err
	}
	return storage.ListJSONNames(dir)
}

func ListReceiverNames() ([]string, error) {
	dir, err := storage.ReceiversDir()
	if err != nil {
		return nil, err
	}
	return storage.ListJSONNames(dir)
}

func RemoveSender(name string) (string, error) {
	path, err := storage.ProfilePath(name)
	if err != nil {
		return "", err
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return path, nil
}
