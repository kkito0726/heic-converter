package encoder

import (
	"image"
	"io"

	"golang.org/x/image/bmp"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/domain/port"
)

// BMPは画像をBMP形式でエンコードする。
type BMP struct{}

var _ port.ImageEncoder = (*BMP)(nil)

// NewBMPはBMPエンコーダを返す。
func NewBMP() *BMP { return &BMP{} }

// Formatはport.ImageEncoderを実装する。
func (e *BMP) Format() model.Format { return model.FormatBMP }

// Encodeはport.ImageEncoderを実装する。BMPは無圧縮なのでopts.Qualityは
// 無視される。
func (e *BMP) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return bmp.Encode(w, img)
}
