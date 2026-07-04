package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/domain/port"
)

func TestConvertImageMultipleFormats(t *testing.T) {
	c := newTestConverter(newFakeStorage())

	got, err := c.ConvertImage(
		context.Background(),
		strings.NewReader("heic-bytes"),
		[]model.Format{model.FormatJPEG, model.FormatPNG},
		model.NewEncodeOptions(model.DefaultQuality),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d images, want 2", len(got))
	}
	// formats指定と同じ順序で返ること
	if got[0].Format != model.FormatJPEG || got[1].Format != model.FormatPNG {
		t.Errorf("formats = %v, %v; want jpg, png", got[0].Format, got[1].Format)
	}
	if string(got[0].Data) != "encoded-jpg" {
		t.Errorf("jpg data = %q", got[0].Data)
	}
	if string(got[1].Data) != "encoded-png" {
		t.Errorf("png data = %q", got[1].Data)
	}
}

func TestConvertImageDecodeFailure(t *testing.T) {
	c := NewConverter(
		&fakeDecoder{failOn: map[string]bool{"broken": true}},
		[]port.ImageEncoder{&fakeEncoder{format: model.FormatJPEG}},
		newFakeStorage(),
	)

	_, err := c.ConvertImage(
		context.Background(),
		strings.NewReader("broken"),
		[]model.Format{model.FormatJPEG},
		model.NewEncodeOptions(0),
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestConvertImageRequiresFormats(t *testing.T) {
	c := newTestConverter(newFakeStorage())

	_, err := c.ConvertImage(context.Background(), strings.NewReader("x"), nil, model.NewEncodeOptions(0))
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestConvertImageUnknownEncoder(t *testing.T) {
	c := newTestConverter(newFakeStorage()) // jpg/pngのエンコーダしか登録されていない

	_, err := c.ConvertImage(
		context.Background(),
		strings.NewReader("x"),
		[]model.Format{model.FormatWebP},
		model.NewEncodeOptions(0),
	)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestConvertImageEncoderFailure(t *testing.T) {
	c := NewConverter(
		&fakeDecoder{},
		[]port.ImageEncoder{&fakeEncoder{format: model.FormatJPEG, fail: true}},
		newFakeStorage(),
	)

	_, err := c.ConvertImage(
		context.Background(),
		strings.NewReader("x"),
		[]model.Format{model.FormatJPEG},
		model.NewEncodeOptions(0),
	)
	if err == nil {
		t.Fatal("expected error from failing encoder")
	}
	if errors.Is(err, ErrInvalidInput) {
		t.Errorf("encoder failure must not be ErrInvalidInput (input was valid): %v", err)
	}
}

func TestConvertImageCanceledContext(t *testing.T) {
	c := newTestConverter(newFakeStorage())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.ConvertImage(ctx, strings.NewReader("x"), []model.Format{model.FormatJPEG}, model.NewEncodeOptions(0))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
