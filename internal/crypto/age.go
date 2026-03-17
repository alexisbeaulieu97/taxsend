package crypto

import (
	"bufio"
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
	r, err := age.ParseX25519Recipient(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid recipient: %w", err)
	}
	return r, nil
}

func LoadIdentity(path string) (*age.X25519Identity, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ids, err := age.ParseIdentities(bufio.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("invalid identity file: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no identities found in %s", path)
	}
	x, ok := ids[0].(*age.X25519Identity)
	if !ok {
		return nil, fmt.Errorf("unsupported identity type in %s", path)
	}
	return x, nil
}

func EncryptStream(dst io.Writer, src io.Reader, recipient age.Recipient) error {
	w, err := age.Encrypt(dst, recipient)
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
