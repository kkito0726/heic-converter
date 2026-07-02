package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

func TestBuildInput(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		opts    *options
		want    usecase.ConvertInput
		wantErr string
	}{
		{
			name: "full options",
			path: "/pics",
			opts: &options{
				formats:   []string{"jpeg", "PNG"},
				outputDir: "/out",
				recursive: true,
				overwrite: true,
				quality:   80,
			},
			want: usecase.ConvertInput{
				Path:      "/pics",
				Recursive: true,
				Formats:   []model.Format{model.FormatJPEG, model.FormatPNG},
				OutputDir: "/out",
				Overwrite: true,
				Quality:   80,
			},
		},
		{
			name:    "missing path",
			path:    "",
			opts:    &options{formats: []string{"jpg"}},
			wantErr: "path",
		},
		{
			name:    "missing formats",
			path:    "/pics",
			opts:    &options{},
			wantErr: "format",
		},
		{
			name:    "invalid format",
			path:    "/pics",
			opts:    &options{formats: []string{"avif"}},
			wantErr: "unsupported format",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildInput(tt.path, tt.opts)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Path != tt.want.Path || got.Recursive != tt.want.Recursive ||
				got.OutputDir != tt.want.OutputDir || got.Overwrite != tt.want.Overwrite ||
				got.Quality != tt.want.Quality {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
			if len(got.Formats) != len(tt.want.Formats) {
				t.Fatalf("formats = %v, want %v", got.Formats, tt.want.Formats)
			}
			for i := range tt.want.Formats {
				if got.Formats[i] != tt.want.Formats[i] {
					t.Errorf("formats[%d] = %v, want %v", i, got.Formats[i], tt.want.Formats[i])
				}
			}
		})
	}
}

func TestPrintSummary(t *testing.T) {
	var sb strings.Builder
	results := []model.ConversionResult{
		{SourcePath: "a.heic", Outputs: []string{"a.jpg"}},
		{SourcePath: "b.heic", Err: errors.New("boom")},
	}
	failed := printSummary(&sb, results)
	if failed != 1 {
		t.Errorf("failed = %d, want 1", failed)
	}
	if !strings.Contains(sb.String(), "1 succeeded, 1 failed") {
		t.Errorf("summary output = %q", sb.String())
	}
}
