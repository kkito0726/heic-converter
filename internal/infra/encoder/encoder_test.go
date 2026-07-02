package encoder

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"testing"

	_ "image/jpeg"
	_ "image/png"

	"heic-converter/internal/domain/model"
	"heic-converter/internal/domain/port"
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

// roundTrip encodes with enc and decodes with the standard image package,
// verifying the encoded bytes are a valid image of the expected size and
// format.
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

func TestAllContainsUniqueFormats(t *testing.T) {
	seen := map[model.Format]bool{}
	for _, enc := range All() {
		f := enc.Format()
		if seen[f] {
			t.Errorf("duplicate encoder for format %v", f)
		}
		seen[f] = true
	}
	if !seen[model.FormatJPEG] || !seen[model.FormatPNG] {
		t.Errorf("All() is missing core encoders: %v", seen)
	}
}

var _ io.Writer = (*bytes.Buffer)(nil)
