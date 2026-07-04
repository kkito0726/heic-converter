package api

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// loggingInterceptorは各RPCのメソッド・所要時間・結果コードを構造化ログに出す。
func loggingInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			res, err := next(ctx, req)
			attrs := []any{
				slog.String("procedure", req.Spec().Procedure),
				slog.Duration("duration", time.Since(start)),
			}
			if err != nil {
				attrs = append(attrs,
					slog.String("code", connect.CodeOf(err).String()),
					slog.String("error", err.Error()),
				)
				logger.Warn("rpc failed", attrs...)
				return nil, err
			}
			logger.Info("rpc ok", attrs...)
			return res, nil
		}
	}
}
