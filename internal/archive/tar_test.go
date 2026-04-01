package archive

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTarRejectsTraversal(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "../../evil.txt", Mode: 0o644, Size: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("evil")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := ExtractTar(bytes.NewReader(buf.Bytes()), t.TempDir(), false)
	if err == nil {
		t.Fatal("expected traversal error")
	}
}

func TestExtractTarUsesPrivateFilePermissions(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{Name: "T4.txt", Mode: 0o644, Size: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("tax!")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()
	if _, err := ExtractTar(bytes.NewReader(buf.Bytes()), outDir, false); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(outDir, "T4.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("got mode %03o, want %03o", got, want)
	}
}
