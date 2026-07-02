package port

import "io"

// FileStorage abstracts file discovery and I/O so the usecase layer never
// touches the OS filesystem directly.
type FileStorage interface {
	// FindFiles returns the file itself when path points at a regular file,
	// or every file directly inside (recursively when recursive is true)
	// when path points at a directory.
	FindFiles(path string, recursive bool) ([]string, error)
	Open(path string) (io.ReadCloser, error)
	// Create opens path for writing, creating parent directories as needed.
	// When overwrite is false and the file already exists, it returns an
	// error wrapping fs.ErrExist.
	Create(path string, overwrite bool) (io.WriteCloser, error)
}
