package chunk

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"taxsend/internal/fsutil"
)

var partPattern = regexp.MustCompile(`^(.+)\.part([0-9]{3})(\.[A-Za-z0-9]+)$`)

type PartInfo struct {
	BaseName string
	Index    int
	Ext      string
}

type ArtifactInfo struct {
	Chunked       bool
	BaseName      string
	Ext           string
	RequestedPart int
	PartPaths     []string
}

func ParsePartFilename(name string) (PartInfo, bool) {
	matches := partPattern.FindStringSubmatch(name)
	if matches == nil {
		return PartInfo{}, false
	}
	index := 0
	for _, ch := range matches[2] {
		index = index*10 + int(ch-'0')
	}
	if index <= 0 {
		return PartInfo{}, false
	}
	return PartInfo{BaseName: matches[1], Index: index, Ext: matches[3]}, true
}

func Collect(path string) (ArtifactInfo, error) {
	filename := filepath.Base(path)
	part, ok := ParsePartFilename(filename)
	if !ok {
		return ArtifactInfo{Chunked: false, PartPaths: []string{path}}, nil
	}
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ArtifactInfo{}, err
	}
	partMap := map[int]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, ok := ParsePartFilename(entry.Name())
		if !ok || info.BaseName != part.BaseName || info.Ext != part.Ext {
			continue
		}
		if _, exists := partMap[info.Index]; exists {
			return ArtifactInfo{}, fmt.Errorf("duplicate artifact part %03d for %s", info.Index, part.BaseName)
		}
		partMap[info.Index] = filepath.Join(dir, entry.Name())
	}
	if len(partMap) == 0 {
		return ArtifactInfo{}, fmt.Errorf("no artifact parts found for %s", path)
	}
	maxIndex := 0
	for index := range partMap {
		if index > maxIndex {
			maxIndex = index
		}
	}
	paths := make([]string, 0, maxIndex)
	for i := 1; i <= maxIndex; i++ {
		partPath, ok := partMap[i]
		if !ok {
			return ArtifactInfo{}, fmt.Errorf("missing artifact part %03d for %s", i, part.BaseName)
		}
		paths = append(paths, partPath)
	}
	return ArtifactInfo{
		Chunked:       true,
		BaseName:      part.BaseName,
		Ext:           part.Ext,
		RequestedPart: part.Index,
		PartPaths:     paths,
	}, nil
}

func FinalizeEncryptedArtifact(tmpPath, outputDir, basename, ext string, maxPartSize int64, force bool) ([]string, error) {
	if maxPartSize <= 0 {
		return nil, fmt.Errorf("max part size must be greater than zero")
	}
	if err := fsutil.EnsurePrivateDir(outputDir); err != nil {
		return nil, err
	}
	stat, err := os.Stat(tmpPath)
	if err != nil {
		return nil, err
	}
	if stat.Size() <= maxPartSize {
		finalPath := filepath.Join(outputDir, basename+ext)
		if err := fsutil.CommitAtomicForce(tmpPath, finalPath, force); err != nil {
			return nil, err
		}
		return []string{finalPath}, nil
	}

	src, err := os.Open(tmpPath)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	defer os.Remove(tmpPath)

	partCount := int((stat.Size() + maxPartSize - 1) / maxPartSize)
	for partNum := 1; partNum <= partCount; partNum++ {
		partPath := filepath.Join(outputDir, fmt.Sprintf("%s.part%03d%s", basename, partNum, ext))
		if err := fsutil.PrepareOutputPath(partPath, force); err != nil {
			return nil, err
		}
	}

	var out []string
	buf := make([]byte, 32*1024)
	for partNum := 1; ; partNum++ {
		partPath := filepath.Join(outputDir, fmt.Sprintf("%s.part%03d%s", basename, partNum, ext))
		tmpPart, tmpPartPath, err := fsutil.CreateAtomic(partPath, force)
		if err != nil {
			return nil, err
		}
		written, err := io.CopyBuffer(tmpPart, io.LimitReader(src, maxPartSize), buf)
		closeErr := tmpPart.Close()
		if err != nil {
			_ = os.Remove(tmpPartPath)
			return nil, err
		}
		if closeErr != nil {
			_ = os.Remove(tmpPartPath)
			return nil, closeErr
		}
		if written == 0 {
			_ = os.Remove(tmpPartPath)
			break
		}
		if err := fsutil.CommitAtomicForce(tmpPartPath, partPath, force); err != nil {
			_ = os.Remove(tmpPartPath)
			return nil, err
		}
		out = append(out, partPath)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no output parts created")
	}
	return out, nil
}

func OpenJoined(path string) (io.ReadCloser, ArtifactInfo, error) {
	info, err := Collect(path)
	if err != nil {
		return nil, ArtifactInfo{}, err
	}
	if len(info.PartPaths) == 1 {
		f, err := os.Open(info.PartPaths[0])
		if err != nil {
			return nil, ArtifactInfo{}, err
		}
		return f, info, nil
	}

	files := make([]*os.File, 0, len(info.PartPaths))
	readers := make([]io.Reader, 0, len(info.PartPaths))
	for _, partPath := range info.PartPaths {
		f, err := os.Open(partPath)
		if err != nil {
			for _, openFile := range files {
				_ = openFile.Close()
			}
			return nil, ArtifactInfo{}, err
		}
		files = append(files, f)
		readers = append(readers, f)
	}
	return &multiReadCloser{reader: io.MultiReader(readers...), files: files}, info, nil
}

type multiReadCloser struct {
	reader io.Reader
	files  []*os.File
}

func (m *multiReadCloser) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func (m *multiReadCloser) Close() error {
	var firstErr error
	for _, f := range m.files {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
