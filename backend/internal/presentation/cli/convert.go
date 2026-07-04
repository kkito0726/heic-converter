package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/usecase"
)

// isTerminalは対話プロンプトとリッチな進捗UIを実行できるかどうかを返す。
// テストでは差し替え可能。
var isTerminal = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

type options struct {
	formats   []string
	outputDir string
	recursive bool
	overwrite bool
	quality   int
}

// newConvertCmdは画像変換コマンドを作る。サブコマンド指定なしで実行された
// ときの動作にあたるため、これがそのままルートコマンドになる。
func newConvertCmd(conv *usecase.Converter) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:          "heic-converter [path]",
		Short:        "Convert HEIC/HEIF images to common formats",
		Long:         "Convert .heic/.heif images to formats such as jpg and png.\nPass a file to convert it, or a directory to convert every HEIC inside.",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			}
			return run(cmd, path, opts, conv)
		},
	}
	f := cmd.Flags()
	f.StringSliceVarP(&opts.formats, "format", "f", nil, "output formats, comma separated (jpg, png, webp, tiff, bmp, gif)")
	f.StringVarP(&opts.outputDir, "output", "o", "", "output directory (default: next to each source file)")
	f.BoolVarP(&opts.recursive, "recursive", "r", false, "process directories recursively")
	f.BoolVar(&opts.overwrite, "overwrite", false, "overwrite existing output files")
	f.IntVarP(&opts.quality, "quality", "q", model.DefaultQuality, "encode quality 1-100 (jpg/webp)")
	return cmd
}

// buildInputはCLI入力を検証し、usecaseの入力にマッピングする。
func buildInput(path string, opts *options) (usecase.ConvertInput, error) {
	if path == "" {
		return usecase.ConvertInput{}, errors.New("path to a .heic file or a directory is required")
	}
	if len(opts.formats) == 0 {
		return usecase.ConvertInput{}, errors.New("at least one output format is required (--format jpg,png)")
	}
	formats, err := model.ParseFormats(opts.formats)
	if err != nil {
		return usecase.ConvertInput{}, err
	}
	return usecase.ConvertInput{
		Path:      path,
		Recursive: opts.recursive,
		Formats:   formats,
		OutputDir: opts.outputDir,
		Overwrite: opts.overwrite,
		Quality:   opts.quality,
	}, nil
}

func run(cmd *cobra.Command, path string, opts *options, conv *usecase.Converter) error {
	if !isTerminal() {
		return runPlain(cmd, path, opts, conv)
	}
	printLogo(cmd.OutOrStdout())
	in, err := runInteractive(conv, path, opts)
	if err != nil {
		return err
	}
	return runWithTUI(cmd.Context(), conv, in, cmd.OutOrStdout())
}

// runPlainは非TTY実行(パイプ・CIなど)を処理する。フラグのみを使い、
// プレーンテキストで出力する。
func runPlain(cmd *cobra.Command, path string, opts *options, conv *usecase.Converter) error {
	in, err := buildInput(path, opts)
	if err != nil {
		return err
	}
	results, err := conv.Convert(cmd.Context(), in, newTextProgress(cmd.OutOrStdout()))
	if err != nil {
		return err
	}
	failed := printSummary(cmd.OutOrStdout(), results)
	if failed > 0 {
		return fmt.Errorf("%d of %d file(s) failed", failed, len(results))
	}
	return nil
}
