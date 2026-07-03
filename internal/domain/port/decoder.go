package port

import (
	"image"
	"io"
)

// ImageDecoderはエンコード済みの元画像(HEICなど)をピクセルデータに変換する。
// 実装はinfra層に置き、domain層・usecase層に触れずにデコードライブラリを
// 差し替えられるようにする。
type ImageDecoder interface {
	Decode(r io.Reader) (image.Image, error)
	// CanDecodeはpathのファイルがこのデコーダでデコード可能そうかどうかを
	// 返す(通常は拡張子で判定する)。
	CanDecode(path string) bool
}
