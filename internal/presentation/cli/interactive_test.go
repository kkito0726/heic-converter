package cli

import (
	"path/filepath"
	"testing"

	"github.com/kkito0726/heic-converter/internal/domain/model"
)

func mustFormats(t *testing.T, names ...string) []model.Format {
	t.Helper()
	formats, err := model.ParseFormats(names)
	if err != nil {
		t.Fatal(err)
	}
	return formats
}

// When path and all options are provided up front, the interactive flow must
// not prompt at all and just assemble the input.
func TestRunInteractiveWithEverythingProvided(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.heic")
	copyFixture(t, src)

	opts := &options{
		formats:   []string{"jpg", "png"},
		outputDir: filepath.Join(dir, "out"),
		recursive: true,
		overwrite: true,
		quality:   85,
	}
	in, err := runInteractive(newRealConverter(), src, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.Path != src || !in.Recursive || !in.Overwrite || in.Quality != 85 {
		t.Errorf("input = %+v", in)
	}
	if len(in.Formats) != 2 {
		t.Errorf("formats = %v", in.Formats)
	}
}

func TestAskPathWithExistingArgument(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.heic")
	copyFixture(t, src)

	t.Run("file", func(t *testing.T) {
		path, isDir, err := askPath(src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path != src || isDir {
			t.Errorf("path=%q isDir=%v", path, isDir)
		}
	})

	t.Run("directory", func(t *testing.T) {
		_, isDir, err := askPath(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isDir {
			t.Error("expected isDir=true")
		}
	})

	t.Run("missing", func(t *testing.T) {
		if _, _, err := askPath(filepath.Join(dir, "nope.heic")); err == nil {
			t.Error("expected error for missing path")
		}
	})
}

func TestValidatePath(t *testing.T) {
	dir := t.TempDir()
	if err := validatePath(dir); err != nil {
		t.Errorf("existing dir should validate: %v", err)
	}
	if err := validatePath(""); err == nil {
		t.Error("empty path should fail")
	}
	if err := validatePath(filepath.Join(dir, "nope")); err == nil {
		t.Error("missing path should fail")
	}
}

func TestAskFormatsWithPreset(t *testing.T) {
	formats, err := askFormats(newRealConverter(), []string{"jpeg", "webp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []model.Format{model.FormatJPEG, model.FormatWebP}
	for i := range want {
		if formats[i] != want[i] {
			t.Errorf("formats[%d] = %v, want %v", i, formats[i], want[i])
		}
	}
}

func TestAskOptionsAllProvided(t *testing.T) {
	opts := &options{outputDir: "/out", recursive: true, overwrite: true}
	if err := askOptions(true, opts); err != nil {
		t.Fatalf("must not prompt when everything is set: %v", err)
	}
}
