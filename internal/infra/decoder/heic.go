// Package decoder provides ImageDecoder implementations.
package decoder

import (
	"fmt"
	"image"
	"io"
	"path/filepath"
	"strings"

	"github.com/gen2brain/heic"
)

// HEIC decodes .heic/.heif files. The underlying library runs a WASM build
// of the decoder via a pure-Go runtime, so the binary stays free of cgo and
// external dependencies.
type HEIC struct{}

// NewHEIC returns a HEIC decoder.
func NewHEIC() *HEIC { return &HEIC{} }

var heicExtensions = map[string]bool{
	".heic": true,
	".heif": true,
}

// Decode implements port.ImageDecoder.
func (d *HEIC) Decode(r io.Reader) (image.Image, error) {
	img, err := heic.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode heic: %w", err)
	}
	return img, nil
}

// CanDecode implements port.ImageDecoder. It judges by file extension,
// case-insensitively.
func (d *HEIC) CanDecode(path string) bool {
	return heicExtensions[strings.ToLower(filepath.Ext(path))]
}
