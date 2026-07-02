package encoder

import (
	"image"
	"image/png"
	"io"

	"github.com/kkito0726/heic-converter/internal/domain/model"
)

// PNG encodes images as PNG using the standard library.
type PNG struct{}

// NewPNG returns a PNG encoder.
func NewPNG() *PNG { return &PNG{} }

// Format implements port.ImageEncoder.
func (e *PNG) Format() model.Format { return model.FormatPNG }

// Encode implements port.ImageEncoder. PNG is lossless, so opts.Quality is
// ignored.
func (e *PNG) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return png.Encode(w, img)
}
