// Command heic-converter converts HEIC/HEIF images to common formats.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kkito0726/heic-converter/internal/infra/decoder"
	"github.com/kkito0726/heic-converter/internal/infra/encoder"
	"github.com/kkito0726/heic-converter/internal/infra/storage"
	"github.com/kkito0726/heic-converter/internal/presentation/cli"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conv := usecase.NewConverter(decoder.NewHEIC(), encoder.All(), storage.NewLocalFS())
	if err := cli.New(conv).ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
