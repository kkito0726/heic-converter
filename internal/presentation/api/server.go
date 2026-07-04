package api

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/kkito0726/heic-converter/internal/gen/heic/v1/heicv1connect"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

// DefaultMaxRequestBytesはリクエストボディの既定サイズ上限(64MiB)。
const DefaultMaxRequestBytes = 64 << 20

const shutdownTimeout = 10 * time.Second

// Configはサーバーの起動設定。
type Config struct {
	Host string
	Port int
	// MaxRequestBytesはリクエストサイズ上限。0以下ならDefaultMaxRequestBytes。
	MaxRequestBytes int
	// AllowedOriginsはブラウザ(connect-web)からの別オリジンアクセスを許可する
	// オリジンの一覧。空ならCORSヘッダを付与しない(同一オリジン配信や
	// 非ブラウザクライアントのみの構成)。
	AllowedOrigins []string
	// Loggerはnilなら標準エラーへのテキストロガー。
	Logger *slog.Logger
}

func (c Config) maxRequestBytes() int {
	if c.MaxRequestBytes <= 0 {
		return DefaultMaxRequestBytes
	}
	return c.MaxRequestBytes
}

func (c Config) logger() *slog.Logger {
	if c.Logger == nil {
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return c.Logger
}

func (c Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// NewHandlerは変換API・ヘルスチェック・リフレクションを束ねたhttp.Handlerを返す。
// gRPCをTLSなしで受けられるようh2c(平文HTTP/2)でラップ済み。
func NewHandler(conv *usecase.Converter, cfg Config) http.Handler {
	mux := http.NewServeMux()

	opts := []connect.HandlerOption{
		connect.WithInterceptors(loggingInterceptor(cfg.logger())),
		connect.WithReadMaxBytes(cfg.maxRequestBytes()),
	}
	mux.Handle(heicv1connect.NewConvertServiceHandler(newHandler(conv), opts...))

	checker := grpchealth.NewStaticChecker(heicv1connect.ConvertServiceName)
	mux.Handle(grpchealth.NewHandler(checker))

	reflector := grpcreflect.NewStaticReflector(heicv1connect.ConvertServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return h2c.NewHandler(withCORS(mux, cfg.AllowedOrigins), &http2.Server{})
}

// withCORSは許可オリジンからのブラウザアクセスに必要なCORSヘッダを付与する。
// 許可・公開するメソッドとヘッダはconnect-rpcの3プロトコルが必要とするものに
// 合わせる(connectrpc.com/corsが提供する一覧を使う)。
func withCORS(h http.Handler, origins []string) http.Handler {
	if len(origins) == 0 {
		return h
	}
	middleware := cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200, // プリフライト結果をブラウザにキャッシュさせる(秒)
	})
	return middleware.Handler(h)
}

// Serveはサーバーを起動し、ctxのキャンセルでgraceful shutdownする。
func Serve(ctx context.Context, conv *usecase.Converter, cfg Config) error {
	logger := cfg.logger()
	srv := &http.Server{
		Addr:              cfg.addr(),
		Handler:           NewHandler(conv, cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	logger.Info("server started",
		slog.String("addr", cfg.addr()),
		slog.String("convert", heicv1connect.ConvertServiceConvertProcedure),
		slog.String("list_formats", heicv1connect.ConvertServiceListFormatsProcedure),
	)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutting down gracefully")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
