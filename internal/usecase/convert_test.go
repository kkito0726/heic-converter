package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"heic-converter/internal/domain/model"
	"heic-converter/internal/domain/port"
)

// --- fakes ---------------------------------------------------------------

type fakeDecoder struct {
	failOn map[string]bool // source basenames that fail to decode
}

func (d *fakeDecoder) Decode(r io.Reader) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if d.failOn[string(data)] {
		return nil, errors.New("corrupt image")
	}
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}

func (d *fakeDecoder) CanDecode(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".heic" || ext == ".heif"
}

type fakeEncoder struct {
	format model.Format
	fail   bool
}

func (e *fakeEncoder) Format() model.Format { return e.format }

func (e *fakeEncoder) Encode(w io.Writer, _ image.Image, _ model.EncodeOptions) error {
	if e.fail {
		return errors.New("encoder broken")
	}
	_, err := w.Write([]byte("encoded-" + string(e.format)))
	return err
}

// fakeStorage is an in-memory FileStorage. files maps path -> content.
type fakeStorage struct {
	mu    sync.Mutex
	files map[string]string
	dirs  map[string][]string // dir path -> contained file paths
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{files: map[string]string{}, dirs: map[string][]string{}}
}

func (s *fakeStorage) addFile(path, content string) {
	s.files[path] = content
	dir := filepath.Dir(path)
	s.dirs[dir] = append(s.dirs[dir], path)
}

func (s *fakeStorage) FindFiles(path string, _ bool) ([]string, error) {
	if files, ok := s.dirs[path]; ok {
		return files, nil
	}
	if _, ok := s.files[path]; ok {
		return []string{path}, nil
	}
	return nil, fmt.Errorf("stat %s: %w", path, fs.ErrNotExist)
}

func (s *fakeStorage) Open(path string) (io.ReadCloser, error) {
	content, ok := s.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader([]byte(content))), nil
}

type writeCloser struct {
	bytes.Buffer
	onClose func(string)
}

func (w *writeCloser) Close() error {
	w.onClose(w.String())
	return nil
}

func (s *fakeStorage) Create(path string, overwrite bool) (io.WriteCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.files[path]; exists && !overwrite {
		return nil, fmt.Errorf("create %s: %w", path, fs.ErrExist)
	}
	s.files[path] = ""
	return &writeCloser{onClose: func(content string) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.files[path] = content
	}}, nil
}

// recordingObserver captures progress callbacks.
type recordingObserver struct {
	total int
	dones []int
}

func (o *recordingObserver) OnStart(total int) { o.total = total }
func (o *recordingObserver) OnFileDone(_ model.ConversionResult, done, _ int) {
	o.dones = append(o.dones, done)
}

// --- helpers ---------------------------------------------------------------

func newTestConverter(st *fakeStorage, encoders ...port.ImageEncoder) *Converter {
	if len(encoders) == 0 {
		encoders = []port.ImageEncoder{
			&fakeEncoder{format: model.FormatJPEG},
			&fakeEncoder{format: model.FormatPNG},
		}
	}
	return NewConverter(&fakeDecoder{}, encoders, st)
}

// --- tests -----------------------------------------------------------------

func TestConvertSingleFileMultipleFormats(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/a.heic", "a")
	c := newTestConverter(st)
	obs := &recordingObserver{}

	results, err := c.Convert(context.Background(), ConvertInput{
		Path:    "/pics/a.heic",
		Formats: []model.Format{model.FormatJPEG, model.FormatPNG},
	}, obs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Succeeded() {
		t.Fatalf("results = %+v", results)
	}
	if len(results[0].Outputs) != 2 {
		t.Fatalf("outputs = %v, want 2", results[0].Outputs)
	}
	if got := st.files["/pics/a.jpg"]; got != "encoded-jpg" {
		t.Errorf("jpg content = %q", got)
	}
	if got := st.files["/pics/a.png"]; got != "encoded-png" {
		t.Errorf("png content = %q", got)
	}
	if obs.total != 1 || len(obs.dones) != 1 {
		t.Errorf("observer: total=%d dones=%v", obs.total, obs.dones)
	}
}

func TestConvertDirectorySkipsNonImages(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/a.heic", "a")
	st.addFile("/pics/b.HEIC", "b")
	st.addFile("/pics/notes.txt", "n")
	c := newTestConverter(st)

	results, err := c.Convert(context.Background(), ConvertInput{
		Path:    "/pics",
		Formats: []model.Format{model.FormatJPEG},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("converted %d files, want 2: %+v", len(results), results)
	}
}

func TestConvertFailSoft(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/good.heic", "good")
	st.addFile("/pics/bad.heic", "bad")
	c := NewConverter(
		&fakeDecoder{failOn: map[string]bool{"bad": true}},
		[]port.ImageEncoder{&fakeEncoder{format: model.FormatJPEG}},
		st,
	)

	results, err := c.Convert(context.Background(), ConvertInput{
		Path:    "/pics",
		Formats: []model.Format{model.FormatJPEG},
	}, nil)
	if err != nil {
		t.Fatalf("run must not abort on a single failure: %v", err)
	}
	var ok, failed int
	for _, r := range results {
		if r.Succeeded() {
			ok++
		} else {
			failed++
		}
	}
	if ok != 1 || failed != 1 {
		t.Errorf("ok=%d failed=%d, want 1/1: %+v", ok, failed, results)
	}
}

func TestConvertRespectsOverwriteFlag(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/a.heic", "a")
	st.addFile("/pics/a.jpg", "existing")
	c := newTestConverter(st)

	results, err := c.Convert(context.Background(), ConvertInput{
		Path:    "/pics/a.heic",
		Formats: []model.Format{model.FormatJPEG},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Succeeded() {
		t.Fatal("expected failure without --overwrite")
	}
	if !errors.Is(results[0].Err, fs.ErrExist) {
		t.Errorf("err = %v, want fs.ErrExist", results[0].Err)
	}
	if st.files["/pics/a.jpg"] != "existing" {
		t.Errorf("existing file was clobbered: %q", st.files["/pics/a.jpg"])
	}
}

func TestConvertNoSources(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/notes.txt", "n")
	c := newTestConverter(st)

	_, err := c.Convert(context.Background(), ConvertInput{
		Path:    "/pics",
		Formats: []model.Format{model.FormatJPEG},
	}, nil)
	if !errors.Is(err, ErrNoSourceFiles) {
		t.Errorf("err = %v, want ErrNoSourceFiles", err)
	}
}

func TestConvertUnknownEncoder(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/a.heic", "a")
	c := newTestConverter(st) // only jpg/png encoders registered

	_, err := c.Convert(context.Background(), ConvertInput{
		Path:    "/pics/a.heic",
		Formats: []model.Format{model.FormatWebP},
	}, nil)
	if err == nil {
		t.Fatal("expected error for missing encoder")
	}
}

func TestConvertRequiresFormats(t *testing.T) {
	st := newFakeStorage()
	st.addFile("/pics/a.heic", "a")
	c := newTestConverter(st)

	if _, err := c.Convert(context.Background(), ConvertInput{Path: "/pics/a.heic"}, nil); err == nil {
		t.Fatal("expected error for empty formats")
	}
}

func TestSupportedFormats(t *testing.T) {
	c := newTestConverter(newFakeStorage())
	got := c.SupportedFormats()
	want := []model.Format{model.FormatJPEG, model.FormatPNG}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		outputDir string
		format    model.Format
		want      string
	}{
		{"same dir", "/pics/a.heic", "", model.FormatJPEG, "/pics/a.jpg"},
		{"custom dir", "/pics/a.heic", "/out", model.FormatPNG, "/out/a.png"},
		{"no extension", "/pics/raw", "", model.FormatJPEG, "/pics/raw.jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OutputPath(tt.src, tt.outputDir, tt.format); got != tt.want {
				t.Errorf("OutputPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
