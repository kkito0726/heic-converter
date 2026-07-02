// Package usecase implements application logic on top of domain ports.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/domain/port"
)

// ErrNoSourceFiles is returned when the given path contains nothing the
// decoder can handle.
var ErrNoSourceFiles = errors.New("no convertible image files found")

// ProgressObserver receives conversion progress. Implementations are called
// from a single goroutine (calls are serialized by the converter).
type ProgressObserver interface {
	OnStart(totalFiles int)
	OnFileDone(result model.ConversionResult, done, total int)
}

type nopObserver struct{}

func (nopObserver) OnStart(int)                                {}
func (nopObserver) OnFileDone(model.ConversionResult, int, int) {}

// ConvertInput carries everything needed for one conversion run.
type ConvertInput struct {
	// Path is a source file or a directory containing source files.
	Path      string
	Recursive bool
	// Formats are the output formats; every source is encoded to all of them.
	Formats []model.Format
	// OutputDir is where outputs are written. Empty means next to each source.
	OutputDir string
	Overwrite bool
	Quality   int
	// Parallelism limits concurrent conversions. <=0 means NumCPU.
	Parallelism int
}

// Converter is the conversion usecase. All I/O goes through the injected
// ports, keeping this layer library-agnostic.
type Converter struct {
	decoder  port.ImageDecoder
	encoders map[model.Format]port.ImageEncoder
	storage  port.FileStorage
}

// NewConverter wires a Converter from port implementations.
func NewConverter(decoder port.ImageDecoder, encoders []port.ImageEncoder, storage port.FileStorage) *Converter {
	byFormat := make(map[model.Format]port.ImageEncoder, len(encoders))
	for _, e := range encoders {
		byFormat[e.Format()] = e
	}
	return &Converter{decoder: decoder, encoders: byFormat, storage: storage}
}

// SupportedFormats returns the formats this converter can encode to, in the
// canonical display order.
func (c *Converter) SupportedFormats() []model.Format {
	var formats []model.Format
	for _, f := range model.AllFormats() {
		if _, ok := c.encoders[f]; ok {
			formats = append(formats, f)
		}
	}
	return formats
}

// FindSources lists the decodable files under path.
func (c *Converter) FindSources(path string, recursive bool) ([]string, error) {
	files, err := c.storage.FindFiles(path, recursive)
	if err != nil {
		return nil, err
	}
	var sources []string
	for _, f := range files {
		if c.decoder.CanDecode(f) {
			sources = append(sources, f)
		}
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("%w under %s", ErrNoSourceFiles, path)
	}
	return sources, nil
}

// Convert converts every decodable file under in.Path to all requested
// formats. Individual file failures do not abort the run; they are recorded
// in the returned results (fail-soft). The returned error is reserved for
// setup problems (bad path, unknown format, cancellation).
func (c *Converter) Convert(ctx context.Context, in ConvertInput, obs ProgressObserver) ([]model.ConversionResult, error) {
	if obs == nil {
		obs = nopObserver{}
	}
	if err := c.validateFormats(in.Formats); err != nil {
		return nil, err
	}
	sources, err := c.FindSources(in.Path, in.Recursive)
	if err != nil {
		return nil, err
	}
	obs.OnStart(len(sources))

	results := make([]model.ConversionResult, len(sources))
	var done atomic.Int64
	var mu sync.Mutex // serializes observer callbacks
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(parallelism(in.Parallelism))
	for i, src := range sources {
		g.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			res := c.convertOne(src, in)
			results[i] = res
			mu.Lock()
			obs.OnFileDone(res, int(done.Add(1)), len(sources))
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *Converter) validateFormats(formats []model.Format) error {
	if len(formats) == 0 {
		return errors.New("at least one output format is required")
	}
	for _, f := range formats {
		if _, ok := c.encoders[f]; !ok {
			return fmt.Errorf("no encoder available for format %q", f)
		}
	}
	return nil
}

func parallelism(n int) int {
	if n <= 0 {
		return runtime.NumCPU()
	}
	return n
}

// convertOne decodes src once and encodes it to every requested format.
func (c *Converter) convertOne(src string, in ConvertInput) model.ConversionResult {
	img, err := c.decode(src)
	if err != nil {
		return model.ConversionResult{SourcePath: src, Err: err}
	}
	opts := model.NewEncodeOptions(in.Quality)
	var outputs []string
	var errs []error
	for _, format := range in.Formats {
		out := OutputPath(src, in.OutputDir, format)
		if err := c.encodeTo(out, img, format, in.Overwrite, opts); err != nil {
			errs = append(errs, err)
			continue
		}
		outputs = append(outputs, out)
	}
	return model.ConversionResult{SourcePath: src, Outputs: outputs, Err: errors.Join(errs...)}
}

func (c *Converter) decode(src string) (image.Image, error) {
	f, err := c.storage.Open(src)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", src, err)
	}
	defer f.Close()
	img, err := c.decoder.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", src, err)
	}
	return img, nil
}

func (c *Converter) encodeTo(path string, img image.Image, format model.Format, overwrite bool, opts model.EncodeOptions) error {
	w, err := c.storage.Create(path, overwrite)
	if err != nil {
		return err
	}
	if err := c.encoders[format].Encode(w, img, opts); err != nil {
		_ = w.Close()
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// OutputPath computes the destination path for a source file and format.
// An empty outputDir places the output next to the source file.
func OutputPath(src, outputDir string, format model.Format) string {
	dir := filepath.Dir(src)
	if outputDir != "" {
		dir = outputDir
	}
	base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	return filepath.Join(dir, base+format.Extension())
}
