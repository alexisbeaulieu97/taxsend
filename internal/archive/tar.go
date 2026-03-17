package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"taxsend/internal/fsutil"
)

type InputFile struct {
	Source string
	TarRel string
	Info   fs.FileInfo
}

func CollectInputs(paths []string) ([]InputFile, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no input files were provided")
	}
	var out []InputFile
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("input path %q: %w", p, err)
		}
		base := filepath.Base(filepath.Clean(p))
		err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			relToRoot, err := filepath.Rel(filepath.Clean(p), path)
			if err != nil {
				return err
			}
			tarRel := base
			if relToRoot != "." {
				tarRel = filepath.ToSlash(filepath.Join(base, relToRoot))
			}
			out = append(out, InputFile{Source: path, TarRel: tarRel, Info: info})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TarRel < out[j].TarRel })
	return out, nil
}

func WriteTar(w io.Writer, files []InputFile) error {
	tw := tar.NewWriter(w)
	defer tw.Close()
	for _, f := range files {
		h, err := tar.FileInfoHeader(f.Info, "")
		if err != nil {
			return err
		}
		h.Name = filepath.ToSlash(f.TarRel)
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		fd, err := os.Open(f.Source)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, fd); err != nil {
			_ = fd.Close()
			return err
		}
		_ = fd.Close()
	}
	return nil
}

func ExtractTar(r io.Reader, outDir string, force bool) (int, error) {
	tr := tar.NewReader(r)
	count := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return count, nil
		}
		if err != nil {
			return count, err
		}
		if h.FileInfo().IsDir() {
			continue
		}
		target, err := fsutil.SafeJoin(outDir, filepath.FromSlash(strings.TrimSpace(h.Name)))
		if err != nil {
			return count, err
		}
		if err := fsutil.EnsureNotExists(target, force); err != nil {
			return count, err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return count, err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, h.FileInfo().Mode().Perm())
		if err != nil {
			return count, err
		}
		if _, err := io.Copy(f, tr); err != nil {
			_ = f.Close()
			return count, err
		}
		if err := f.Close(); err != nil {
			return count, err
		}
		count++
	}
}
