package api

import (
	"bytes"
	"context"
	"image"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"connectrpc.com/connect"

	heicv1 "github.com/kkito0726/heic-converter/backend/internal/gen/heic/v1"
	"github.com/kkito0726/heic-converter/backend/internal/gen/heic/v1/heicv1connect"
	"github.com/kkito0726/heic-converter/backend/internal/infra/decoder"
	"github.com/kkito0726/heic-converter/backend/internal/infra/encoder"
	"github.com/kkito0726/heic-converter/backend/internal/infra/storage"
	"github.com/kkito0726/heic-converter/backend/internal/usecase"
)

const fixtureHEIC = "../../infra/decoder/testdata/sample.heic"

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newRealConverter() *usecase.Converter {
	return usecase.NewConverter(decoder.NewHEIC(), encoder.All(), storage.NewLocalFS())
}

// newTestServerはHTTP/2対応のテストサーバーを起動する(gRPCプロトコルに必要)。
func newTestServer(t *testing.T, cfg Config) *httptest.Server {
	t.Helper()
	cfg.Logger = quietLogger()
	srv := httptest.NewUnstartedServer(NewHandler(newRealConverter(), cfg))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv
}

func readFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(fixtureHEIC)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}

// 3プロトコルすべてで実HEICの変換ができること。
func TestConvertAcrossProtocols(t *testing.T) {
	srv := newTestServer(t, Config{})
	heicData := readFixture(t)

	protocols := []struct {
		name string
		opts []connect.ClientOption
	}{
		{name: "connect", opts: nil},
		{name: "grpc", opts: []connect.ClientOption{connect.WithGRPC()}},
		{name: "grpc-web", opts: []connect.ClientOption{connect.WithGRPCWeb()}},
	}
	for _, p := range protocols {
		t.Run(p.name, func(t *testing.T) {
			client := heicv1connect.NewConvertServiceClient(srv.Client(), srv.URL, p.opts...)
			res, err := client.Convert(context.Background(), connect.NewRequest(&heicv1.ConvertRequest{
				Image:   heicData,
				Formats: []string{"jpg", "png"},
				Quality: 90,
			}))
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}
			images := res.Msg.GetImages()
			if len(images) != 2 {
				t.Fatalf("got %d images, want 2", len(images))
			}
			// formats指定と同じ順序で、デコード可能な画像が返ること
			wantFormats := []string{"jpeg", "png"}
			for i, img := range images {
				decoded, format, err := image.Decode(bytes.NewReader(img.GetData()))
				if err != nil {
					t.Fatalf("images[%d] is not a valid image: %v", i, err)
				}
				if format != wantFormats[i] {
					t.Errorf("images[%d] format = %q, want %q", i, format, wantFormats[i])
				}
				if decoded.Bounds().Dx() != 64 || decoded.Bounds().Dy() != 48 {
					t.Errorf("images[%d] size = %v, want 64x48", i, decoded.Bounds())
				}
			}
		})
	}
}

func TestConvertInvalidArgument(t *testing.T) {
	srv := newTestServer(t, Config{})
	client := heicv1connect.NewConvertServiceClient(srv.Client(), srv.URL)
	heicData := readFixture(t)

	tests := []struct {
		name string
		req  *heicv1.ConvertRequest
	}{
		{name: "empty image", req: &heicv1.ConvertRequest{Formats: []string{"jpg"}}},
		{name: "no formats", req: &heicv1.ConvertRequest{Image: heicData}},
		{name: "unsupported format", req: &heicv1.ConvertRequest{Image: heicData, Formats: []string{"avif"}}},
		{name: "corrupt image", req: &heicv1.ConvertRequest{Image: []byte("not a heic"), Formats: []string{"jpg"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Convert(context.Background(), connect.NewRequest(tt.req))
			if err == nil {
				t.Fatal("expected error")
			}
			if code := connect.CodeOf(err); code != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want invalid_argument (err: %v)", code, err)
			}
		})
	}
}

func TestConvertRequestTooLarge(t *testing.T) {
	srv := newTestServer(t, Config{MaxRequestBytes: 16})
	client := heicv1connect.NewConvertServiceClient(srv.Client(), srv.URL)

	_, err := client.Convert(context.Background(), connect.NewRequest(&heicv1.ConvertRequest{
		Image:   readFixture(t),
		Formats: []string{"jpg"},
	}))
	if err == nil {
		t.Fatal("expected error for oversized request")
	}
	if code := connect.CodeOf(err); code != connect.CodeResourceExhausted {
		t.Errorf("code = %v, want resource_exhausted (err: %v)", code, err)
	}
}

func TestListFormats(t *testing.T) {
	srv := newTestServer(t, Config{})
	client := heicv1connect.NewConvertServiceClient(srv.Client(), srv.URL)

	res, err := client.ListFormats(context.Background(), connect.NewRequest(&heicv1.ListFormatsRequest{}))
	if err != nil {
		t.Fatalf("ListFormats() error: %v", err)
	}
	formats := res.Msg.GetFormats()
	if len(formats) != 6 {
		t.Errorf("got %d formats %v, want 6", len(formats), formats)
	}
}

// 副作用なしメソッドはConnectプロトコルのHTTP GETでも呼べること。
func TestListFormatsWithHTTPGet(t *testing.T) {
	srv := newTestServer(t, Config{})
	url := srv.URL + "/heic.v1.ConvertService/ListFormats?connect=v1&encoding=json&message=%7B%7D"

	res, err := srv.Client().Get(url)
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.StatusCode, body)
	}
	if !strings.Contains(string(body), "jpg") {
		t.Errorf("body missing formats: %s", body)
	}
}

func TestHealthCheck(t *testing.T) {
	srv := newTestServer(t, Config{})

	res, err := srv.Client().Post(
		srv.URL+"/grpc.health.v1.Health/Check",
		"application/json",
		strings.NewReader(`{}`),
	)
	if err != nil {
		t.Fatalf("health check error: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.StatusCode, body)
	}
	if !strings.Contains(string(body), "SERVING") {
		t.Errorf("body = %s, want SERVING", body)
	}
}

// AllowedOrigins設定時、ブラウザ(connect-web)からの別オリジンアクセスに
// 必要なCORSヘッダが返ること。
func TestCORS(t *testing.T) {
	srv := newTestServer(t, Config{AllowedOrigins: []string{"https://front.example.com"}})
	procedure := srv.URL + "/heic.v1.ConvertService/Convert"

	t.Run("preflight from allowed origin", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, procedure, nil)
		req.Header.Set("Origin", "https://front.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		// ブラウザはfetch仕様に従い小文字・空白なしのカンマ区切りで送る
		req.Header.Set("Access-Control-Request-Headers", "connect-protocol-version,content-type")

		res, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("preflight error: %v", err)
		}
		defer res.Body.Close()
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "https://front.example.com" {
			t.Errorf("Access-Control-Allow-Origin = %q, want allowed origin", got)
		}
		if got := res.Header.Get("Access-Control-Allow-Headers"); got == "" {
			t.Error("Access-Control-Allow-Headers is empty")
		}
	})

	t.Run("preflight from disallowed origin", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, procedure, nil)
		req.Header.Set("Origin", "https://evil.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")

		res, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("preflight error: %v", err)
		}
		defer res.Body.Close()
		if got := res.Header.Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want empty for disallowed origin", got)
		}
	})

	t.Run("rpc still works with CORS enabled", func(t *testing.T) {
		client := heicv1connect.NewConvertServiceClient(srv.Client(), srv.URL)
		if _, err := client.ListFormats(context.Background(), connect.NewRequest(&heicv1.ListFormatsRequest{})); err != nil {
			t.Fatalf("ListFormats() error with CORS enabled: %v", err)
		}
	})
}

// AllowedOrigins未設定(デフォルト)ではCORSヘッダを返さないこと。
func TestCORSDisabledByDefault(t *testing.T) {
	srv := newTestServer(t, Config{})
	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/heic.v1.ConvertService/Convert", nil)
	req.Header.Set("Origin", "https://front.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("preflight error: %v", err)
	}
	defer res.Body.Close()
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty when CORS is disabled", got)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	if got := cfg.maxRequestBytes(); got != DefaultMaxRequestBytes {
		t.Errorf("maxRequestBytes() = %d, want %d", got, DefaultMaxRequestBytes)
	}
	if got := (Config{MaxRequestBytes: 128}).maxRequestBytes(); got != 128 {
		t.Errorf("maxRequestBytes() = %d, want 128", got)
	}
	if cfg.logger() == nil {
		t.Error("logger() must not return nil")
	}
	if got := (Config{Host: "127.0.0.1", Port: 8080}).addr(); got != "127.0.0.1:8080" {
		t.Errorf("addr() = %q", got)
	}
}

func TestServeGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, newRealConverter(), Config{
			Host:   "127.0.0.1",
			Port:   0, // 空きポート
			Logger: quietLogger(),
		})
	}()

	// サーバー起動を少し待ってからキャンセルする
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Serve() = %v, want nil after graceful shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve() did not return after context cancellation")
	}
}
