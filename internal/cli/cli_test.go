package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"taxsend/internal/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	inDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(filepath.Join(inDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "T4.txt"), []byte("income"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "nested", "RL1.txt"), []byte("tax"), 0o644); err != nil {
		t.Fatal(err)
	}

	identity := filepath.Join(tmp, "identity.txt")
	encrypted := filepath.Join(tmp, "bundle.tar.age")
	outDir := filepath.Join(tmp, "out")

	root := NewRootCmd()
	root.SetArgs([]string{"keygen", "--output", identity})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	id, err := crypto.LoadIdentity(identity)
	if err != nil {
		t.Fatal(err)
	}
	recipient := id.Recipient().String()

	if _, err := os.Stat(filepath.Join(tmp, "bundle.tar")); !os.IsNotExist(err) {
		t.Fatalf("unexpected plaintext bundle file")
	}

	if _, err := runCapture([]string{"encrypt", "--recipient", recipient, "--output", encrypted, inDir}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(encrypted); err != nil {
		t.Fatal(err)
	}

	if _, err := runCapture([]string{"decrypt", "--identity", identity, "--output-dir", outDir, encrypted}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, filepath.Base(inDir), "T4.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "income" {
		t.Fatalf("got %q", string(got))
	}
}

func runCapture(args []string) ([]byte, error) {
	root := NewRootCmd()
	buf := &strings.Builder{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return []byte(buf.String()), err
}
