package model

import (
	"errors"
	"testing"
)

func TestNewEncodeOptions(t *testing.T) {
	tests := []struct {
		name    string
		quality int
		want    int
	}{
		{"valid", 80, 80},
		{"lower bound", 1, 1},
		{"upper bound", 100, 100},
		{"zero falls back", 0, DefaultQuality},
		{"negative falls back", -5, DefaultQuality},
		{"too high falls back", 101, DefaultQuality},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEncodeOptions(tt.quality); got.Quality != tt.want {
				t.Errorf("NewEncodeOptions(%d).Quality = %d, want %d", tt.quality, got.Quality, tt.want)
			}
		})
	}
}

func TestConversionResultSucceeded(t *testing.T) {
	ok := ConversionResult{SourcePath: "a.heic", Outputs: []string{"a.jpg"}}
	if !ok.Succeeded() {
		t.Error("result without error should succeed")
	}
	bad := ConversionResult{SourcePath: "b.heic", Err: errors.New("boom")}
	if bad.Succeeded() {
		t.Error("result with error should not succeed")
	}
}
