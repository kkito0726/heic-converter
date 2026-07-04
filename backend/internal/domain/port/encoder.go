package port

import (
	"image"
	"io"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
)

// ImageEncoderはピクセルデータを単一の出力形式で書き出す。
type ImageEncoder interface {
	Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error
	// Formatはこのエンコーダがどの出力形式を生成するかを返す。
	Format() model.Format
}
