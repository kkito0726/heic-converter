package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/usecase"
)

// runInteractiveはフラグで指定されなかった項目をユーザーに尋ねることで
// 変換の入力を完成させる。
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

// askPathは引数で与えられていなければ変換元パスを尋ね、
// それがディレクトリかどうかを返す。
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

// askFormatsは--formatが指定されていなければ出力形式の複数選択UIを表示する。
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

// askOptionsはゼロ値のままになっているオプションを尋ねる。
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
