package storage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.heic"))
	writeFile(t, filepath.Join(dir, "b.txt"))
	writeFile(t, filepath.Join(dir, "sub", "c.heic"))
	s := NewLocalFS()

	t.Run("single file", func(t *testing.T) {
		got, err := s.FindFiles(filepath.Join(dir, "a.heic"), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0] != filepath.Join(dir, "a.heic") {
			t.Errorf("got %v", got)
		}
	})

	t.Run("directory non-recursive", func(t *testing.T) {
		got, err := s.FindFiles(dir, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d files %v, want 2 (sub dir excluded)", len(got), got)
		}
	})

	t.Run("directory recursive", func(t *testing.T) {
		got, err := s.FindFiles(dir, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("got %d files %v, want 3", len(got), got)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		if _, err := s.FindFiles(filepath.Join(dir, "nope"), false); err == nil {
			t.Error("expected error for missing path")
		}
	})
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	s := NewLocalFS()
	target := filepath.Join(dir, "out", "img.jpg")

	t.Run("creates parent directories", func(t *testing.T) {
		w, err := s.Create(target, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := w.Write([]byte("data")); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("refuses overwrite by default", func(t *testing.T) {
		_, err := s.Create(target, false)
		if !errors.Is(err, fs.ErrExist) {
			t.Errorf("error = %v, want fs.ErrExist", err)
		}
	})

	t.Run("overwrites when allowed", func(t *testing.T) {
		w, err := s.Create(target, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.heic")
	writeFile(t, path)
	s := NewLocalFS()
	r, err := s.Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}
