package encoder

import (
	"image"
	"io"

	"github.com/gen2brain/webp"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/domain/port"
)

// WebPは画像をWebP形式でエンコードする。HEICデコーダと同様、内部で使う
// ライブラリはWASMビルドをpure Goランタイムで実行するため、cgoは不要。
type WebP struct{}

var _ port.ImageEncoder = (*WebP)(nil)

// NewWebPはWebPエンコーダを返す。
func NewWebP() *WebP { return &WebP{} }

// Formatはport.ImageEncoderを実装する。
func (e *WebP) Format() model.Format { return model.FormatWebP }

// Encodeはport.ImageEncoderを実装する。
func (e *WebP) Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error {
	return webp.Encode(w, img, webp.Options{Quality: opts.Quality})
}
