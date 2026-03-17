package archive

import (
	"archive/tar"
	"bytes"
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
