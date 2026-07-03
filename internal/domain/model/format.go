package model

import (
	"fmt"
	"strings"
)

// Formatはコンバータが対応する出力画像形式を表す。
type Format string

const (
	FormatJPEG Format = "jpg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatTIFF Format = "tiff"
	FormatBMP  Format = "bmp"
	FormatGIF  Format = "gif"
)

// AllFormatsは表示順に、対応するすべての出力形式を返す。
func AllFormats() []Format {
	return []Format{FormatJPEG, FormatPNG, FormatWebP, FormatTIFF, FormatBMP, FormatGIF}
}

var formatAliases = map[string]Format{
	"jpg":  FormatJPEG,
	"jpeg": FormatJPEG,
	"png":  FormatPNG,
	"webp": FormatWebP,
	"tiff": FormatTIFF,
	"tif":  FormatTIFF,
	"bmp":  FormatBMP,
	"gif":  FormatGIF,
}

// ParseFormatは"jpg"、"JPEG"、".png"のようなユーザー入力文字列を
// Formatに変換する。
func ParseFormat(s string) (Format, error) {
	key := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), "."))
	f, ok := formatAliases[key]
	if !ok {
		return "", fmt.Errorf("unsupported format %q (supported: %s)", s, supportedNames())
	}
	return f, nil
}

// ParseFormatsは複数の形式文字列をパースする。元の順序を保ったまま
// 重複は除去する。
func ParseFormats(ss []string) ([]Format, error) {
	seen := make(map[Format]bool, len(ss))
	formats := make([]Format, 0, len(ss))
	for _, s := range ss {
		f, err := ParseFormat(s)
		if err != nil {
			return nil, err
		}
		if seen[f] {
			continue
		}
		seen[f] = true
		formats = append(formats, f)
	}
	return formats, nil
}

func supportedNames() string {
	names := make([]string, 0, len(AllFormats()))
	for _, f := range AllFormats() {
		names = append(names, string(f))
	}
	return strings.Join(names, ", ")
}

func (f Format) String() string { return string(f) }

// Extensionは先頭のドットを含む、その形式のファイル拡張子を返す。
func (f Format) Extension() string { return "." + string(f) }
