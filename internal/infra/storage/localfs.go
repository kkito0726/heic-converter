// Package storage provides FileStorage implementations.
package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// LocalFS implements port.FileStorage against the local filesystem.
type LocalFS struct{}

// NewLocalFS returns a LocalFS storage.
func NewLocalFS() *LocalFS { return &LocalFS{} }

// FindFiles implements port.FileStorage.
func (s *LocalFS) FindFiles(path string, recursive bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	files, err := listDir(path, recursive)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", path, err)
	}
	sort.Strings(files)
	return files, nil
}

func listDir(dir string, recursive bool) ([]string, error) {
	if !recursive {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		var files []string
		for _, e := range entries {
			if !e.IsDir() {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
		return files, nil
	}
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

// Open implements port.FileStorage.
func (s *LocalFS) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Create implements port.FileStorage. When overwrite is false and the file
// already exists, the returned error wraps fs.ErrExist.
func (s *LocalFS) Create(path string, overwrite bool) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	return f, nil
}
