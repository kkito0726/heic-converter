package encoder

import (
	"image"
	"image/gif"
	"io"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/domain/port"
)

// GIF encodes images as GIF using the standard library. Colors are reduced
// to a 256-color palette as required by the format.
type GIF struct{}

var _ port.ImageEncoder = (*GIF)(nil)

// NewGIF returns a GIF encoder.
func NewGIF() *GIF { return &GIF{} }

// Format implements port.ImageEncoder.
func (e *GIF) Format() model.Format { return model.FormatGIF }

// Encode implements port.ImageEncoder. GIF palette quantization ignores
// opts.Quality.
func (e *GIF) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return gif.Encode(w, img, &gif.Options{NumColors: 256})
}
