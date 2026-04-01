package cli

import (
	"io"
	"os"
	"path/filepath"
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
	root.SetArgs(args)

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		_ = stdoutR.Close()
		_ = stdoutW.Close()
		return nil, err
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutW, stderrW
	root.SetOut(stdoutW)
	root.SetErr(stderrW)
	err = root.Execute()
	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout, os.Stderr = oldStdout, oldStderr

	stdoutData, readErr := io.ReadAll(stdoutR)
	if readErr != nil {
		return nil, readErr
	}
	stderrData, readErr := io.ReadAll(stderrR)
	if readErr != nil {
		return nil, readErr
	}
	_ = stdoutR.Close()
	_ = stderrR.Close()
	return append(stdoutData, stderrData...), err
}
