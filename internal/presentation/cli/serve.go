package cli

import (
	"github.com/spf13/cobra"

	"github.com/kkito0726/heic-converter/internal/presentation/api"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

// newServeCmdはWebサーバーモード(gRPC / gRPC-Web / Connect)を起動する
// serveサブコマンドを作る。
func newServeCmd(conv *usecase.Converter) *cobra.Command {
	cfg := api.Config{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the conversion web server (gRPC / gRPC-Web / Connect)",
		Long:  "Start a web server that converts HEIC/HEIF images.\nA single endpoint accepts gRPC, gRPC-Web and Connect (HTTP+JSON) requests.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return api.Serve(cmd.Context(), conv, cfg)
		},
	}
	f := cmd.Flags()
	f.StringVar(&cfg.Host, "host", "0.0.0.0", "address to bind")
	f.IntVar(&cfg.Port, "port", 8080, "port to listen on")
	f.IntVar(&cfg.MaxRequestBytes, "max-request-bytes", api.DefaultMaxRequestBytes, "request body size limit in bytes")
	f.StringSliceVar(&cfg.AllowedOrigins, "allowed-origins", nil, "CORS origins allowed for browser clients (e.g. http://localhost:5173)")
	return cmd
}
