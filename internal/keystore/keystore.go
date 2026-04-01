package keystore

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"taxsend/internal/crypto"
	"taxsend/internal/fsutil"
	"taxsend/internal/storage"
)

func SaveIdentity(name string, identity *age.X25519Identity, passphrase string, force bool) (string, error) {
	if identity == nil {
		return "", fmt.Errorf("identity is required")
	}
	path, err := storage.ReceiverKeystorePath(name)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := crypto.EncryptWithPassphrase(buf, bytes.NewBufferString(identity.String()+"\n"), passphrase); err != nil {
		return "", err
	}
	return path, fsutil.WriteFileAtomic(path, buf.Bytes(), 0o600, force)
}

func SaveMigratedIdentity(name, legacyPath, passphrase string, force bool) (*age.X25519Identity, string, error) {
	identity, err := crypto.LoadIdentity(legacyPath)
	if err != nil {
		return nil, "", err
	}
	path, err := SaveIdentity(name, identity, passphrase, force)
	if err != nil {
		return nil, "", err
	}
	return identity, path, nil
}

func LoadIdentity(name, passphrase string) (*age.X25519Identity, error) {
	path, err := storage.ReceiverKeystorePath(name)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader, err := crypto.DecryptWithPassphrase(f, passphrase)
	if err != nil {
		return nil, fmt.Errorf("open receiver keystore: %w", err)
	}
	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("decrypt receiver keystore: %w", err)
	}
	return crypto.ParseIdentityBytes(plaintext)
}
