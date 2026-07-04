package model

import (
	"testing"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Format
		wantErr bool
	}{
		{name: "jpg", in: "jpg", want: FormatJPEG},
		{name: "jpeg alias", in: "jpeg", want: FormatJPEG},
		{name: "uppercase", in: "PNG", want: FormatPNG},
		{name: "leading dot", in: ".webp", want: FormatWebP},
		{name: "surrounding spaces", in: " gif ", want: FormatGIF},
		{name: "tif alias", in: "tif", want: FormatTIFF},
		{name: "bmp", in: "bmp", want: FormatBMP},
		{name: "unsupported", in: "avif", wantErr: true},
		{name: "empty", in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFormat(%q) expected error, got %v", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFormat(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseFormats(t *testing.T) {
	t.Run("deduplicates keeping order", func(t *testing.T) {
		got, err := ParseFormats([]string{"jpeg", "png", "jpg", "PNG"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []Format{FormatJPEG, FormatPNG}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %v, want %v", i, got[i], want[i])
			}
		}
	})

	t.Run("propagates parse error", func(t *testing.T) {
		if _, err := ParseFormats([]string{"jpg", "nope"}); err == nil {
			t.Fatal("expected error for unsupported format")
		}
	})
}

func TestFormatExtension(t *testing.T) {
	if got := FormatJPEG.Extension(); got != ".jpg" {
		t.Errorf("Extension() = %q, want %q", got, ".jpg")
	}
	if got := FormatWebP.String(); got != "webp" {
		t.Errorf("String() = %q, want %q", got, "webp")
	}
}
