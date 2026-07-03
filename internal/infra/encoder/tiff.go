package encoder

import (
	"image"
	"io"

	"golang.org/x/image/tiff"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/domain/port"
)

// TIFFはdeflate圧縮で画像をTIFF形式にエンコードする。
type TIFF struct{}

var _ port.ImageEncoder = (*TIFF)(nil)

// NewTIFFはTIFFエンコーダを返す。
func NewTIFF() *TIFF { return &TIFF{} }

// Formatはport.ImageEncoderを実装する。
func (e *TIFF) Format() model.Format { return model.FormatTIFF }

// Encodeはport.ImageEncoderを実装する。TIFFは可逆形式なのでopts.Qualityは
// 無視される。
func (e *TIFF) Encode(w io.Writer, img image.Image, _ model.EncodeOptions) error {
	return tiff.Encode(w, img, &tiff.Options{Compression: tiff.Deflate})
}
