// encoderパッケージは各出力形式向けのImageEncoder実装を提供する。
package encoder

import (
	"image"
	"image/jpeg"
	"io"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/domain/port"
)

// JPEGは標準ライブラリを使って画像をJPEG形式でエンコードする。
type JPEG struct{}

var _ port.ImageEncoder = (*JPEG)(nil)

// NewJPEGはJPEGエンコーダを返す。
func NewJPEG() *JPEG { return &JPEG{} }

// Formatはport.ImageEncoderを実装する。
func (e *JPEG) Format() model.Format { return model.FormatJPEG }

// Encodeはport.ImageEncoderを実装する。
func (e *JPEG) Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: opts.Quality})
}
