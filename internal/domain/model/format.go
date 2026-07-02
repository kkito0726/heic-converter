package model

import (
	"fmt"
	"strings"
)

// Format is an output image format supported by the converter.
type Format string

const (
	FormatJPEG Format = "jpg"
	FormatPNG  Format = "png"
	FormatWebP Format = "webp"
	FormatTIFF Format = "tiff"
	FormatBMP  Format = "bmp"
	FormatGIF  Format = "gif"
)

// AllFormats returns every supported output format in display order.
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

// ParseFormat converts a user-supplied string such as "jpg", "JPEG" or ".png"
// into a Format.
func ParseFormat(s string) (Format, error) {
	key := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), "."))
	f, ok := formatAliases[key]
	if !ok {
		return "", fmt.Errorf("unsupported format %q (supported: %s)", s, supportedNames())
	}
	return f, nil
}

// ParseFormats parses multiple format strings, deduplicating while keeping
// the original order.
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

// Extension returns the file extension for the format, including the leading dot.
func (f Format) Extension() string { return "." + string(f) }
