// decoderパッケージはImageDecoderの実装を提供する。
package decoder

import (
	"fmt"
	"image"
	"io"
	"path/filepath"
	"strings"

	"github.com/gen2brain/heic"

	"github.com/kkito0726/heic-converter/internal/domain/port"
)

// HEICは.heic/.heifファイルをデコードする。内部で使うライブラリは
// デコーダをWASMビルドしたものをpure Goランタイムで実行するため、
// バイナリはcgoや外部依存から解放される。
type HEIC struct{}

var _ port.ImageDecoder = (*HEIC)(nil)

// NewHEICはHEICデコーダを返す。
func NewHEIC() *HEIC { return &HEIC{} }

var heicExtensions = map[string]bool{
	".heic": true,
	".heif": true,
}

// Decodeはport.ImageDecoderを実装する。
func (d *HEIC) Decode(r io.Reader) (image.Image, error) {
	img, err := heic.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode heic: %w", err)
	}
	return img, nil
}

// CanDecodeはport.ImageDecoderを実装する。大文字小文字を区別せず、
// 拡張子で判定する。
func (d *HEIC) CanDecode(path string) bool {
	return heicExtensions[strings.ToLower(filepath.Ext(path))]
}
