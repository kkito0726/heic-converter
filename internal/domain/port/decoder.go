package port

import (
	"image"
	"io"
)

// ImageDecoder turns an encoded source image (e.g. HEIC) into pixel data.
// Implementations live in the infra layer so the decoding library can be
// swapped without touching domain or usecase code.
type ImageDecoder interface {
	Decode(r io.Reader) (image.Image, error)
	// CanDecode reports whether the file at path looks decodable by this
	// decoder (typically judged by file extension).
	CanDecode(path string) bool
}
