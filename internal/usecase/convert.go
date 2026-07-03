// usecaseパッケージはdomain層のportの上にアプリケーションロジックを実装する。
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

// ErrNoSourceFilesは指定パス配下にデコーダが扱えるファイルが1つもない
// 場合に返される。
var ErrNoSourceFiles = errors.New("no convertible image files found")

// ProgressObserverは変換の進捗を受け取る。実装への呼び出しは単一の
// ゴルーチンから行われる(converterが呼び出しを直列化する)。
type ProgressObserver interface {
	OnStart(totalFiles int)
	OnFileDone(result model.ConversionResult, done, total int)
}

type nopObserver struct{}

var _ ProgressObserver = nopObserver{}

func (nopObserver) OnStart(int)                                {}
func (nopObserver) OnFileDone(model.ConversionResult, int, int) {}

// ConvertInputは1回の変換実行に必要な情報をすべて保持する。
type ConvertInput struct {
	// Pathは変換元ファイル、または変換元ファイルを含むディレクトリ。
	Path      string
	Recursive bool
	// Formatsは出力形式。すべての変換元ファイルがこれら全形式にエンコード
	// される。
	Formats []model.Format
	// OutputDirは出力先。空の場合は各変換元ファイルと同じ場所に出力する。
	OutputDir string
	Overwrite bool
	Quality   int
	// Parallelismは同時変換数を制限する。0以下ならNumCPUを使う。
	Parallelism int
}

// Converterは変換のusecase。すべてのI/Oは注入されたportを経由するため、
// この層は特定のライブラリに依存しない。
type Converter struct {
	decoder  port.ImageDecoder
	encoders map[model.Format]port.ImageEncoder
	storage  port.FileStorage
}

// NewConverterはportの実装からConverterを組み立てる。
func NewConverter(decoder port.ImageDecoder, encoders []port.ImageEncoder, storage port.FileStorage) *Converter {
	byFormat := make(map[model.Format]port.ImageEncoder, len(encoders))
	for _, e := range encoders {
		byFormat[e.Format()] = e
	}
	return &Converter{decoder: decoder, encoders: byFormat, storage: storage}
}

// SupportedFormatsはこのconverterがエンコード可能な形式を、標準の表示順で
// 返す。
func (c *Converter) SupportedFormats() []model.Format {
	var formats []model.Format
	for _, f := range model.AllFormats() {
		if _, ok := c.encoders[f]; ok {
			formats = append(formats, f)
		}
	}
	return formats
}

// FindSourcesはpath配下のデコード可能なファイルを列挙する。
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

// Convertはin.Path配下のデコード可能なファイルすべてを、要求されたすべての
// 形式に変換する。個々のファイルの失敗は実行全体を中断せず、戻り値の
// resultsに記録される(fail-soft)。戻り値のerrorはセットアップ上の問題
// (不正なパス、未知の形式、キャンセル)専用に予約されている。
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
	var mu sync.Mutex // observerへのコールバックを直列化する
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

// convertOneはsrcを1回だけデコードし、要求されたすべての形式にエンコード
// する。
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

// OutputPathは変換元ファイルと形式から出力先パスを計算する。outputDirが
// 空の場合、出力は変換元ファイルと同じ場所に置かれる。
func OutputPath(src, outputDir string, format model.Format) string {
	dir := filepath.Dir(src)
	if outputDir != "" {
		dir = outputDir
	}
	base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))
	return filepath.Join(dir, base+format.Extension())
}
