package encoder

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"testing"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gen2brain/webp"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/domain/port"
)

func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 16, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 16), G: uint8(y * 20), B: 128, A: 255})
		}
	}
	return img
}

// roundTripはencでエンコードし、標準のimageパッケージでデコードすることで、
// エンコードされたバイト列が期待するサイズ・形式の妥当な画像であることを
// 検証する。
func roundTrip(t *testing.T, enc port.ImageEncoder, wantFormat string) {
	t.Helper()
	var buf bytes.Buffer
	src := testImage()
	if err := enc.Encode(&buf, src, model.NewEncodeOptions(model.DefaultQuality)); err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decoding %s output: %v", wantFormat, err)
	}
	if format != wantFormat {
		t.Errorf("decoded format = %q, want %q", format, wantFormat)
	}
	if decoded.Bounds() != src.Bounds() {
		t.Errorf("decoded bounds = %v, want %v", decoded.Bounds(), src.Bounds())
	}
}

func TestJPEGEncoder(t *testing.T) {
	enc := NewJPEG()
	if enc.Format() != model.FormatJPEG {
		t.Errorf("Format() = %v, want %v", enc.Format(), model.FormatJPEG)
	}
	roundTrip(t, enc, "jpeg")
}

func TestPNGEncoder(t *testing.T) {
	enc := NewPNG()
	if enc.Format() != model.FormatPNG {
		t.Errorf("Format() = %v, want %v", enc.Format(), model.FormatPNG)
	}
	roundTrip(t, enc, "png")
}

func TestTIFFEncoder(t *testing.T) {
	enc := NewTIFF()
	if enc.Format() != model.FormatTIFF {
		t.Errorf("Format() = %v, want %v", enc.Format(), model.FormatTIFF)
	}
	roundTrip(t, enc, "tiff")
}

func TestBMPEncoder(t *testing.T) {
	enc := NewBMP()
	if enc.Format() != model.FormatBMP {
		t.Errorf("Format() = %v, want %v", enc.Format(), model.FormatBMP)
	}
	roundTrip(t, enc, "bmp")
}

func TestGIFEncoder(t *testing.T) {
	enc := NewGIF()
	if enc.Format() != model.FormatGIF {
		t.Errorf("Format() = %v, want %v", enc.Format(), model.FormatGIF)
	}
	roundTrip(t, enc, "gif")
}

// 標準ライブラリはwebpに対応していないため、WebP出力はwebpパッケージ自身で
// デコードする。
func TestWebPEncoder(t *testing.T) {
	enc := NewWebP()
	if enc.Format() != model.FormatWebP {
		t.Errorf("Format() = %v, want %v", enc.Format(), model.FormatWebP)
	}
	var buf bytes.Buffer
	src := testImage()
	if err := enc.Encode(&buf, src, model.NewEncodeOptions(model.DefaultQuality)); err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	decoded, err := webp.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decoding webp output: %v", err)
	}
	if decoded.Bounds() != src.Bounds() {
		t.Errorf("decoded bounds = %v, want %v", decoded.Bounds(), src.Bounds())
	}
}

func TestAllContainsUniqueFormats(t *testing.T) {
	seen := map[model.Format]bool{}
	for _, enc := range All() {
		f := enc.Format()
		if seen[f] {
			t.Errorf("duplicate encoder for format %v", f)
		}
		seen[f] = true
	}
	for _, f := range model.AllFormats() {
		if !seen[f] {
			t.Errorf("All() is missing an encoder for %v", f)
		}
	}
}

var _ io.Writer = (*bytes.Buffer)(nil)
