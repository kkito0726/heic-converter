package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/kkito0726/heic-converter/internal/domain/model"
)

// ErrInvalidInputは呼び出し側の入力不備(未対応の形式・壊れた画像など)による
// 失敗を表す。presentation層はこれを4xx系(invalid_argument)にマッピングできる。
var ErrInvalidInput = errors.New("invalid input")

// ConvertedImageは1形式分の変換結果。
type ConvertedImage struct {
	Format model.Format
	Data   []byte
}

// ConvertImageはrから画像を1回だけデコードし、指定された全形式にエンコードして
// formatsと同じ順序で返す。ファイルシステムには一切触れない(FileStorage不使用)。
// APIのような同期呼び出し向けのため、Convertと違いfail-softにはせず、
// どれか1つでも失敗したら全体をエラーにする。
func (c *Converter) ConvertImage(ctx context.Context, r io.Reader, formats []model.Format, opts model.EncodeOptions) ([]ConvertedImage, error) {
	if err := c.validateFormats(formats); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	img, err := c.decoder.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("%w: decode image: %v", ErrInvalidInput, err)
	}
	results := make([]ConvertedImage, 0, len(formats))
	for _, format := range formats {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := c.encoders[format].Encode(&buf, img, opts); err != nil {
			return nil, fmt.Errorf("encode %s: %w", format, err)
		}
		results = append(results, ConvertedImage{Format: format, Data: buf.Bytes()})
	}
	return results, nil
}
