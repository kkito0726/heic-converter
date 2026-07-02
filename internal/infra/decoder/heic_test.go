package decoder

import (
	"os"
	"testing"
)

func TestHEICDecode(t *testing.T) {
	f, err := os.Open("testdata/sample.heic")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	d := NewHEIC()
	img, err := d.Decode(f)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 48 {
		t.Errorf("decoded size = %dx%d, want 64x48", bounds.Dx(), bounds.Dy())
	}
}

func TestHEICDecodeInvalidData(t *testing.T) {
	f, err := os.Open("heic.go") // not an image
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := NewHEIC().Decode(f); err == nil {
		t.Error("expected error for non-HEIC data")
	}
}

func TestHEICCanDecode(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"a.heic", true},
		{"a.HEIC", true},
		{"b.heif", true},
		{"b.HeIf", true},
		{"c.jpg", false},
		{"noext", false},
		{"dir.heic/file.txt", false},
	}
	d := NewHEIC()
	for _, tt := range tests {
		if got := d.CanDecode(tt.path); got != tt.want {
			t.Errorf("CanDecode(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
