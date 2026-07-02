package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kkito0726/heic-converter/internal/infra/decoder"
	"github.com/kkito0726/heic-converter/internal/infra/encoder"
	"github.com/kkito0726/heic-converter/internal/infra/storage"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

const fixtureHEIC = "../../infra/decoder/testdata/sample.heic"

func newRealConverter() *usecase.Converter {
	return usecase.NewConverter(decoder.NewHEIC(), encoder.All(), storage.NewLocalFS())
}

func copyFixture(t *testing.T, dst string) {
	t.Helper()
	data, err := os.ReadFile(fixtureHEIC)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func forcePlainMode(t *testing.T) {
	t.Helper()
	old := isTerminal
	isTerminal = func() bool { return false }
	t.Cleanup(func() { isTerminal = old })
}

func TestRootCommandPlainRun(t *testing.T) {
	forcePlainMode(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "a.heic")
	copyFixture(t, src)

	cmd := New(newRealConverter())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{src, "--format", "jpg,png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error: %v\noutput: %s", err, out.String())
	}
	for _, name := range []string{"a.jpg", "a.png"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected output %s: %v", name, err)
		}
	}
	if !strings.Contains(out.String(), "1 succeeded, 0 failed") {
		t.Errorf("summary missing: %q", out.String())
	}
}

func TestRootCommandPlainRequiresFlags(t *testing.T) {
	forcePlainMode(t)
	cmd := New(newRealConverter())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error without path in non-TTY mode")
	}
}

func TestRunWithTUIHeadless(t *testing.T) {
	old := extraProgramOptions
	extraProgramOptions = []tea.ProgramOption{
		tea.WithoutRenderer(),
		tea.WithInput(strings.NewReader("")),
	}
	t.Cleanup(func() { extraProgramOptions = old })

	dir := t.TempDir()
	src := filepath.Join(dir, "a.heic")
	copyFixture(t, src)

	in := usecase.ConvertInput{
		Path:    src,
		Formats: mustFormats(t, "jpg"),
	}
	var out bytes.Buffer
	if err := runWithTUI(context.Background(), newRealConverter(), in, &out); err != nil {
		t.Fatalf("runWithTUI() error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.jpg")); err != nil {
		t.Errorf("expected output a.jpg: %v", err)
	}
	if !strings.Contains(out.String(), "1 succeeded") {
		t.Errorf("summary missing: %q", out.String())
	}
}
