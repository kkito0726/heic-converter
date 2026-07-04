package encoder

import (
	"image"
	"image/gif"
	"io"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/domain/port"
)

// GIFは標準ライブラリを使って画像をGIF形式でエンコードする。
// 形式の仕様上、色は256色パレットに減色される。
type GIF struct{}

var _ port.ImageEncoder = (*GIF)(nil)

// NewGIFはGIFエンコーダを返す。
func NewGIF() *GIF { return &GIF{} }

// Formatはport.ImageEncoderを実装する。
func (e *GIF) Format() model.Format { return model.FormatGIF }

// Encodeはport.ImageEncoderを実装する。GIFのパレット量子化では
// opts.Qualityは無視される。
func (e *GIF) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return gif.Encode(w, img, &gif.Options{NumColors: 256})
}
