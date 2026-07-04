package encoder

import "github.com/kkito0726/heic-converter/backend/internal/domain/port"

// Allは利用可能なエンコーダ実装をすべて返す。
func All() []port.ImageEncoder {
	return []port.ImageEncoder{
		NewJPEG(),
		NewPNG(),
		NewWebP(),
		NewTIFF(),
		NewBMP(),
		NewGIF(),
	}
}
