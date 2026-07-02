package encoder

import (
	"image"
	"io"

	"github.com/gen2brain/webp"

	"heic-converter/internal/domain/model"
)

// WebP encodes images as WebP. Like the HEIC decoder, the underlying library
// runs a WASM build via a pure-Go runtime, so no cgo is involved.
type WebP struct{}

// NewWebP returns a WebP encoder.
func NewWebP() *WebP { return &WebP{} }

// Format implements port.ImageEncoder.
func (e *WebP) Format() model.Format { return model.FormatWebP }

// Encode implements port.ImageEncoder.
func (e *WebP) Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error {
	return webp.Encode(w, img, webp.Options{Quality: opts.Quality})
}
