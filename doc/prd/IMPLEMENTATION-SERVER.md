# 実装方針: Webサーバーモード(connect-rpc)

[PRD-SERVER.md](./PRD-SERVER.md)を実現するための技術選定・設計・実装計画。
既存の[IMPLEMENTATION.md](./IMPLEMENTATION.md)のアーキテクチャを前提に、presentation層を追加する。

## 1. 技術選定

| 用途 | 選定 | 理由 |
|---|---|---|
| RPCフレームワーク | [connectrpc.com/connect](https://connectrpc.com)(connect-go) | 1つの`http.Handler`でgRPC / gRPC-Web / Connectの3プロトコルを自動判別。pure Goで`CGO_ENABLED=0`を維持できる。`net/http`標準のミドルウェアがそのまま使える |
| スキーマ管理・コード生成 | [buf](https://buf.build)(buf CLI + `protoc-gen-go` + `protoc-gen-connect-go`) | protoのlint・破壊的変更検知・生成を一括管理。protocの手動インストール不要 |
| TLSなしHTTP/2 | `golang.org/x/net/http2/h2c` | gRPCはHTTP/2必須。TLS終端はリバースプロキシ前提のため平文HTTP/2(h2c)で待ち受ける |
| ヘルスチェック | `connectrpc.com/grpchealth` | gRPC Health Checking Protocol準拠 |
| リフレクション | `connectrpc.com/grpcreflect` | grpcurlをprotoなしで使えるようにする |
| ログ | `log/slog`(標準ライブラリ) | 構造化ログ。追加依存なし |

**gRPC公式(google.golang.org/grpc)ではなくconnect-goを選ぶ理由**: 公式gRPCサーバーはgRPCプロトコル専用で、REST的なJSON受付には別途grpc-gatewayのリバースプロキシ層が必要になる。connect-goは単一ハンドラで3プロトコルを受けられ、構成要素が少ない。

## 2. API定義(proto)

`proto/heic/v1/convert.proto` に定義し、bufで生成する。

```proto
syntax = "proto3";

package heic.v1;

// HEIC画像の変換サービス
service ConvertService {
  // HEIC/HEIF画像を指定された形式に変換する
  rpc Convert(ConvertRequest) returns (ConvertResponse) {}

  // 対応する出力形式の一覧を返す(副作用なし: ConnectプロトコルではGET可)
  rpc ListFormats(ListFormatsRequest) returns (ListFormatsResponse) {
    option idempotency_level = NO_SIDE_EFFECTS;
  }
}

message ConvertRequest {
  // HEIC/HEIFのバイナリ(Connect JSONではbase64文字列)
  bytes image = 1;
  // 出力形式: "jpg", "png", "webp", "tiff", "bmp", "gif"(複数指定可)
  repeated string formats = 2;
  // 非可逆形式の品質1-100。0または省略時はデフォルト(90)
  uint32 quality = 3;
}

message ConvertedImage {
  string format = 1;
  bytes data = 2;
}

message ConvertResponse {
  // formats指定と同じ順序で返す
  repeated ConvertedImage images = 1;
}

message ListFormatsRequest {}

message ListFormatsResponse {
  repeated string formats = 1;
}
```

- 生成コードは`internal/gen/heic/v1/`に出力(`buf.gen.yaml`で指定)し、**リポジトリにコミットする**(利用側が`go install`だけでビルドできるようにするため)
- `buf lint`と`buf generate`の差分チェックをCIに追加し、protoと生成コードの乖離を防ぐ

## 3. アーキテクチャ

### 3.1 usecase層の拡張(バイト列ベースの変換)

既存の`Converter.Convert`はファイルパス+`FileStorage`前提のため、サーバーからは使いにくい。
**入出力を`io.Reader`/バイト列で扱う変換メソッドをusecaseに追加**し、デコーダ・エンコーダのportをそのまま再利用する。

```go
// ConvertedImageは1形式分の変換結果。
type ConvertedImage struct {
    Format model.Format
    Data   []byte
}

// ConvertImageはrから画像を1回デコードし、指定された全形式にエンコードして返す。
// ファイルシステムには一切触れない(FileStorageを使わない)。
func (c *Converter) ConvertImage(
    ctx context.Context,
    r io.Reader,
    formats []model.Format,
    opts model.EncodeOptions,
) ([]ConvertedImage, error)
```

- 検証ロジック(`validateFormats`)・「1回デコード・複数エンコード」の方針はCLI経路と共通
- 既存のパスベース`Convert`はそのまま残す(CLIの挙動に変更なし)。内部の共通化は挙動を変えない範囲でリファクタリングする

### 3.2 presentation/api層(新規)

設計当初から予約していた`internal/presentation/api/`に実装する。**usecase層にのみ依存**し、生成コード(`internal/gen`)との型変換を担う。

```
proto/heic/v1/convert.proto     # スキーマ(Single Source of Truth)
buf.yaml / buf.gen.yaml         # buf設定
internal/
  gen/heic/v1/                  # buf生成コード(コミットする)
    convert.pb.go
    heicv1connect/convert.connect.go
  presentation/api/
    handler.go                  # ConvertServiceHandler実装(usecaseへ委譲、エラーコード変換)
    server.go                   # http.Server組み立て(h2c・health・reflection・graceful shutdown)
    interceptor.go              # ログ用interceptor(slog)
  presentation/cli/
    serve.go                    # serveサブコマンド(フラグ→api.Serverの起動)
```

### 3.3 リクエストの流れ

```
client (curl / grpcurl / ブラウザ)
  │  Connect(JSON) / gRPC / gRPC-Web
  ▼
http.Server + h2c ── logging interceptor
  ▼
api.handler.Convert(ctx, req)
  │  formats文字列 → model.ParseFormats
  │  quality       → model.NewEncodeOptions
  ▼
usecase.Converter.ConvertImage(ctx, bytes.Reader, formats, opts)
  │  port.ImageDecoder.Decode(1回)
  │  port.ImageEncoder.Encode(形式ごと)
  ▼
api.handler ── []ConvertedImage → ConvertResponse に詰め替えて返却
```

### 3.4 エラーマッピング

| 事象 | Connectコード |
|---|---|
| formatsが空・未対応の形式名 | `invalid_argument` |
| 画像のデコード失敗(壊れたファイル・HEIC以外) | `invalid_argument` |
| リクエストサイズ超過(`connect.WithReadMaxBytes`) | `resource_exhausted` |
| コンテキストキャンセル/デッドライン超過 | `canceled` / `deadline_exceeded` |
| その他の内部エラー | `internal`(詳細はログのみに出し、レスポンスには含めない) |

### 3.5 serveサブコマンド

```go
heic-converter serve --port 8080 --host 0.0.0.0 --max-request-bytes 67108864
```

- cobraのサブコマンドとして`cli`パッケージに追加(既存のルートコマンド動作は不変)
- `signal.NotifyContext`(既存のmainのctx)を受けて`http.Server.Shutdown`でgraceful shutdown
- 起動時にリッスンアドレスとエンドポイント一覧をログ出力する

## 4. 動作確認方法

```sh
# Connectプロトコル(curl + JSON、imageはbase64)
curl -X POST http://localhost:8080/heic.v1.ConvertService/Convert \
  -H "Content-Type: application/json" \
  -d "{\"image\":\"$(base64 -i photo.heic)\",\"formats\":[\"jpg\",\"png\"],\"quality\":90}"

# 形式一覧(副作用なしなのでGET可)
curl "http://localhost:8080/heic.v1.ConvertService/ListFormats?encoding=json&message=%7B%7D"

# gRPC(リフレクション利用)
grpcurl -plaintext -d '{"image":"<base64>","formats":["jpg"]}' \
  localhost:8080 heic.v1.ConvertService/Convert

# ヘルスチェック
grpcurl -plaintext localhost:8080 grpc.health.v1.Health/Check
```

## 5. テスト方針(TDD・カバレッジ80%以上)

| 対象 | 方針 |
|---|---|
| `usecase.ConvertImage` | 既存のfakeデコーダ/エンコーダを再利用したユニットテスト。複数形式・デコード失敗・空formats・キャンセルを網羅 |
| `api.handler` | `httptest.Server` + connect-goクライアントによるin-processテスト。**Connect / gRPC / gRPC-Webの3プロトコルすべて**で実HEICフィクスチャ(`testdata/sample.heic`)の変換を検証 |
| エラー系 | 未対応形式→`invalid_argument`、壊れた画像→`invalid_argument`、サイズ超過→`resource_exhausted`をコードレベルで検証 |
| serveコマンド | フラグ→サーバー設定へのマッピングを関数に切り出して単体テスト。graceful shutdownはテスト用ポートで起動→SIGTERM相当→終了を確認 |
| E2E(手動) | 上記「動作確認方法」のcurl / grpcurlをリリース前に実施 |

## 6. 実装フェーズ

### Phase 1: スキーマとコード生成基盤
1. buf CLIをdevcontainer/CIに追加、`buf.yaml` / `buf.gen.yaml`を作成
2. `proto/heic/v1/convert.proto`定義 → `internal/gen/`へ生成・コミット
3. CIに`buf lint` + 生成コード差分チェックを追加

### Phase 2: usecase拡張
4. `usecase.ConvertImage`をTDDで実装(fakeポートでRED→GREEN)

### Phase 3: サーバー実装
5. `api.handler`(Convert / ListFormats、エラーマッピング)をTDDで実装
6. `api.server`(h2c・ReadMaxBytes・graceful shutdown)実装
7. `serve`サブコマンド追加、実HEICでの3プロトコル疎通テスト

### Phase 4: 運用機能・仕上げ
8. health / reflection / ログinterceptor
9. README・doc/SEQUENCE.mdにサーバーモードを追記
10. (将来)vanguardによる純REST化、streaming RPC

## 7. リスクと対応

| リスク | 対応 |
|---|---|
| WASMデコーダ(wazero)の並行実行時の性能・安全性 | CLIの並列変換(NumCPU並列)で既に実績あり。負荷試験で問題が出た場合はセマフォで同時デコード数を制限する |
| 大きい画像によるメモリ圧迫 | `WithReadMaxBytes`でリクエストサイズを制限(デフォルト64MiB)。1リクエストの処理中のみメモリに保持 |
| Connect JSONのbase64によるペイロード膨張(約1.33倍) | サイズ上限で担保。大容量が必要なクライアントにはbinary(protobuf)エンコーディングを案内 |
| 生成コードとprotoの乖離 | CIで`buf generate`後の`git diff --exit-code`を検証 |
| h2c(平文HTTP/2)の露出 | 本番はリバースプロキシ(TLS終端)配下での運用を前提とし、READMEに明記 |
