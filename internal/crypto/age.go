package crypto

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
)

func GenerateIdentity() (*age.X25519Identity, error) {
	return age.GenerateX25519Identity()
}

func ParseRecipient(s string) (age.Recipient, error) {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "age1pq1"):
		r, err := age.ParseHybridRecipient(s)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
		return r, nil
	case strings.HasPrefix(s, "age1"):
		r, err := age.ParseX25519Recipient(s)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
		return r, nil
	default:
		return nil, fmt.Errorf("unsupported recipient type: %q", s)
	}
}

func LoadIdentity(path string) (*age.X25519Identity, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadIdentityReader(f, path)
}

func LoadIdentityReader(r io.Reader, label string) (*age.X25519Identity, error) {
	ids, err := age.ParseIdentities(bufio.NewReader(r))
	if err != nil {
		return nil, fmt.Errorf("invalid identity file: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no identities found in %s", label)
	}
	x, ok := ids[0].(*age.X25519Identity)
	if !ok {
		return nil, fmt.Errorf("unsupported identity type in %s", label)
	}
	return x, nil
}

func ParseIdentityBytes(data []byte) (*age.X25519Identity, error) {
	return LoadIdentityReader(bytes.NewReader(data), "memory")
}

func ParseRecipients(values []string, recipientsFile string) ([]age.Recipient, error) {
	var recipients []age.Recipient
	for _, value := range values {
		r, err := ParseRecipient(value)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}
	if strings.TrimSpace(recipientsFile) != "" {
		f, err := os.Open(recipientsFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fileRecipients, err := age.ParseRecipients(bufio.NewReader(f))
		if err != nil {
			return nil, fmt.Errorf("invalid recipients file: %w", err)
		}
		recipients = append(recipients, fileRecipients...)
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}
	return recipients, nil
}

func EncryptStream(dst io.Writer, src io.Reader, recipients ...age.Recipient) error {
	w, err := age.Encrypt(dst, recipients...)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, src); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func DecryptStream(src io.Reader, id age.Identity) (io.Reader, error) {
	return age.Decrypt(src, id)
}

func EncryptWithPassphrase(dst io.Writer, src io.Reader, passphrase string) error {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return err
	}
	return EncryptStream(dst, src, recipient)
}

func DecryptWithPassphrase(src io.Reader, passphrase string) (io.Reader, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, err
	}
	return age.Decrypt(src, identity)
}
