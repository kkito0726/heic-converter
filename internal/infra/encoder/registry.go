package encoder

import "heic-converter/internal/domain/port"

// All returns every available encoder implementation.
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
