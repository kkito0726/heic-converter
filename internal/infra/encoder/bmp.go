package encoder

import (
	"image"
	"io"

	"golang.org/x/image/bmp"

	"github.com/kkito0726/heic-converter/internal/domain/model"
)

// BMP encodes images as BMP.
type BMP struct{}

// NewBMP returns a BMP encoder.
func NewBMP() *BMP { return &BMP{} }

// Format implements port.ImageEncoder.
func (e *BMP) Format() model.Format { return model.FormatBMP }

// Encode implements port.ImageEncoder. BMP is uncompressed, so opts.Quality
// is ignored.
func (e *BMP) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return bmp.Encode(w, img)
}
