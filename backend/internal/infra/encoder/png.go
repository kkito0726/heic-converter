package encoder

import (
	"image"
	"image/png"
	"io"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/domain/port"
)

// PNGは標準ライブラリを使って画像をPNG形式でエンコードする。
type PNG struct{}

var _ port.ImageEncoder = (*PNG)(nil)

// NewPNGはPNGエンコーダを返す。
func NewPNG() *PNG { return &PNG{} }

// Formatはport.ImageEncoderを実装する。
func (e *PNG) Format() model.Format { return model.FormatPNG }

// Encodeはport.ImageEncoderを実装する。PNGは可逆形式なのでopts.Qualityは
// 無視される。
func (e *PNG) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return png.Encode(w, img)
}
