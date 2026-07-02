// Package encoder provides ImageEncoder implementations for each output format.
package encoder

import (
	"image"
	"image/jpeg"
	"io"

	"heic-converter/internal/domain/model"
)

// JPEG encodes images as JPEG using the standard library.
type JPEG struct{}

// NewJPEG returns a JPEG encoder.
func NewJPEG() *JPEG { return &JPEG{} }

// Format implements port.ImageEncoder.
func (e *JPEG) Format() model.Format { return model.FormatJPEG }

// Encode implements port.ImageEncoder.
func (e *JPEG) Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: opts.Quality})
}
