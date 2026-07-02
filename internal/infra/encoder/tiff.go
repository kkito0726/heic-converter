package encoder

import (
	"image"
	"io"

	"golang.org/x/image/tiff"

	"heic-converter/internal/domain/model"
)

// TIFF encodes images as TIFF with deflate compression.
type TIFF struct{}

// NewTIFF returns a TIFF encoder.
func NewTIFF() *TIFF { return &TIFF{} }

// Format implements port.ImageEncoder.
func (e *TIFF) Format() model.Format { return model.FormatTIFF }

// Encode implements port.ImageEncoder. TIFF is lossless, so opts.Quality is
// ignored.
func (e *TIFF) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return tiff.Encode(w, img, &tiff.Options{Compression: tiff.Deflate})
}
