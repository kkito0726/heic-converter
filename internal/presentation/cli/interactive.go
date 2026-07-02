package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

// runInteractive completes the conversion input by asking the user for
// whatever the flags did not provide.
func runInteractive(conv *usecase.Converter, path string, opts *options) (usecase.ConvertInput, error) {
	path, isDir, err := askPath(path)
	if err != nil {
		return usecase.ConvertInput{}, err
	}
	formats, err := askFormats(conv, opts.formats)
	if err != nil {
		return usecase.ConvertInput{}, err
	}
	if err := askOptions(isDir, opts); err != nil {
		return usecase.ConvertInput{}, err
	}
	return usecase.ConvertInput{
		Path:      path,
		Recursive: opts.recursive,
		Formats:   formats,
		OutputDir: strings.TrimSpace(opts.outputDir),
		Overwrite: opts.overwrite,
		Quality:   opts.quality,
	}, nil
}

// askPath prompts for the source path unless it was given as an argument,
// and reports whether it is a directory.
func askPath(path string) (string, bool, error) {
	if strings.TrimSpace(path) == "" {
		input := huh.NewInput().
			Title("Path to convert").
			Description("A .heic file, or a directory containing .heic files").
			Placeholder("./photos").
			Value(&path).
			Validate(validatePath)
		if err := huh.NewForm(huh.NewGroup(input)).Run(); err != nil {
			return "", false, err
		}
	}
	path = strings.TrimSpace(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", false, fmt.Errorf("cannot access %s: %w", path, err)
	}
	return path, info.IsDir(), nil
}

func validatePath(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("path is required")
	}
	if _, err := os.Stat(s); err != nil {
		return errors.New("path does not exist")
	}
	return nil
}

// askFormats shows a multi-select of output formats unless --format was given.
func askFormats(conv *usecase.Converter, preset []string) ([]model.Format, error) {
	if len(preset) > 0 {
		return model.ParseFormats(preset)
	}
	supported := conv.SupportedFormats()
	choices := make([]huh.Option[model.Format], 0, len(supported))
	for _, f := range supported {
		choices = append(choices, huh.NewOption(string(f), f))
	}
	var selected []model.Format
	field := huh.NewMultiSelect[model.Format]().
		Title("Output formats").
		Description("Space to toggle, enter to confirm").
		Options(choices...).
		Value(&selected).
		Validate(func(fs []model.Format) error {
			if len(fs) == 0 {
				return errors.New("select at least one format")
			}
			return nil
		})
	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, err
	}
	return selected, nil
}

// askOptions prompts for options that are still at their zero value.
func askOptions(isDir bool, opts *options) error {
	var fields []huh.Field
	if isDir && !opts.recursive {
		fields = append(fields, huh.NewConfirm().
			Title("Include subdirectories?").
			Value(&opts.recursive))
	}
	if opts.outputDir == "" {
		fields = append(fields, huh.NewInput().
			Title("Output directory").
			Description("Leave empty to save next to each source file").
			Value(&opts.outputDir))
	}
	if !opts.overwrite {
		fields = append(fields, huh.NewConfirm().
			Title("Overwrite existing files?").
			Value(&opts.overwrite))
	}
	if len(fields) == 0 {
		return nil
	}
	return huh.NewForm(huh.NewGroup(fields...)).Run()
}
