package port

import (
	"image"
	"io"

	"github.com/kkito0726/heic-converter/internal/domain/model"
)

// ImageEncoder writes pixel data out in a single target format.
type ImageEncoder interface {
	Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error
	// Format identifies which output format this encoder produces.
	Format() model.Format
}
