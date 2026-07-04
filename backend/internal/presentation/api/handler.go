// apiパッケージはconnect-rpcによるWebサーバーのpresentation層。
// usecase層にのみ依存し、gRPC / gRPC-Web / Connectの3プロトコルを受け付ける。
package api

import (
	"bytes"
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	heicv1 "github.com/kkito0726/heic-converter/backend/internal/gen/heic/v1"
	"github.com/kkito0726/heic-converter/backend/internal/gen/heic/v1/heicv1connect"
	"github.com/kkito0726/heic-converter/backend/internal/usecase"
)

// handlerはConvertServiceの実装。usecaseへの委譲と型変換のみを担う。
type handler struct {
	conv *usecase.Converter
}

var _ heicv1connect.ConvertServiceHandler = (*handler)(nil)

func newHandler(conv *usecase.Converter) *handler {
	return &handler{conv: conv}
}

// ConvertはHEIC/HEIF画像を指定された全形式に変換して返す。
func (h *handler) Convert(ctx context.Context, req *connect.Request[heicv1.ConvertRequest]) (*connect.Response[heicv1.ConvertResponse], error) {
	msg := req.Msg
	if len(msg.GetImage()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("image is required"))
	}
	formats, err := model.ParseFormats(msg.GetFormats())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	opts := model.NewEncodeOptions(int(msg.GetQuality()))

	images, err := h.conv.ConvertImage(ctx, bytes.NewReader(msg.GetImage()), formats, opts)
	if err != nil {
		return nil, toConnectError(err)
	}

	res := &heicv1.ConvertResponse{
		Images: make([]*heicv1.ConvertedImage, 0, len(images)),
	}
	for _, img := range images {
		res.Images = append(res.Images, &heicv1.ConvertedImage{
			Format: string(img.Format),
			Data:   img.Data,
		})
	}
	return connect.NewResponse(res), nil
}

// ListFormatsは対応する出力形式の一覧を返す。
func (h *handler) ListFormats(_ context.Context, _ *connect.Request[heicv1.ListFormatsRequest]) (*connect.Response[heicv1.ListFormatsResponse], error) {
	formats := h.conv.SupportedFormats()
	names := make([]string, 0, len(formats))
	for _, f := range formats {
		names = append(names, string(f))
	}
	return connect.NewResponse(&heicv1.ListFormatsResponse{Formats: names}), nil
}

// toConnectErrorはusecaseのエラーをConnectのエラーコードにマッピングする。
func toConnectError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, context.Canceled):
		return connect.NewError(connect.CodeCanceled, err)
	case errors.Is(err, context.DeadlineExceeded):
		return connect.NewError(connect.CodeDeadlineExceeded, err)
	default:
		// 内部詳細はレスポンスに載せない(ログにはinterceptorが出す)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}
