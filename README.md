# heic-converter

[![CI](https://github.com/kkito0726/heic-converter/actions/workflows/ci.yml/badge.svg)](https://github.com/kkito0726/heic-converter/actions/workflows/ci.yml)

`.heic` / `.heif` 画像を jpg / png / webp / tiff / bmp / gif に変換する、対話型のCLIツールです。

- **単一バイナリ** — cgo不使用(`CGO_ENABLED=0`)。ImageMagickやlibheifなどのインストールは一切不要で、バイナリ1つを置くだけで macOS / Linux / Windows で動作します
- **対話モード** — 引数なしで起動すると、パス入力 → チェックボックスで出力形式を複数選択 → オプション確認、と対話形式で迷わず変換できます
- **一括変換** — フォルダを渡せば配下のHEICをまとめて並列変換(サブフォルダの再帰処理にも対応)
- **リッチなUI** — アスキーアートのロゴ、カラーテーマ、プログレスバーでの進捗表示
- **fail-soft** — 1ファイルの失敗で全体は止まらず、最後に失敗理由をまとめてレポートします

## HEICファイルとは?

HEIC (High Efficiency Image Container) は、**HEIF** (High Efficiency Image File Format) という規格に基づく画像フォーマットで、動画コーデック HEVC (H.265) の技術を静止画圧縮に応用したものです。

- **iPhone / iPad の写真はデフォルトでHEIC形式**で保存されます(iOS 11以降)
- JPEGの**約半分のファイルサイズ**で同等の画質を保てるのが利点です
- 一方で、**Apple製品以外との互換性が低い**のが難点です。Windowsの古い環境・多くのWebサービス・一部の画像ビューアではそのまま開けず、アップロードを拒否されることもあります

このツールは、そうしたHEICファイルを互換性の高いjpgやpngなどにローカルで変換します。オンラインコンバータと違い**写真をどこにもアップロードしない**ので、プライバシーの面でも安心です。

## インストール

### インストールスクリプト(推奨)

macOS / Linux / Windows (Git Bash) 対応。OSとCPUアーキテクチャを自動判別して最新リリースをインストールします。

```sh
curl -fsSL https://raw.githubusercontent.com/kkito0726/heic-converter/main/install.sh | sh
```

バージョンやインストール先は環境変数で指定できます。

```sh
VERSION=v0.1.0 BIN_DIR=~/bin curl -fsSL https://raw.githubusercontent.com/kkito0726/heic-converter/main/install.sh | sh
```

### Goでインストール

```sh
go install github.com/kkito0726/heic-converter/backend/cmd/heic-converter@latest
```

### ソースからビルド

```sh
git clone https://github.com/kkito0726/heic-converter.git
cd heic-converter/backend
CGO_ENABLED=0 go build -o heic-converter ./cmd/heic-converter
```

## 使い方

### 対話モード

引数なし(またはパスだけ)で起動すると対話形式で進みます。

```sh
heic-converter
# または
heic-converter ./photos
```

1. 変換したいファイルまたはフォルダのパスを入力
2. 出力形式をチェックボックスで選択(スペースで選択、複数可)
3. 出力先・上書き可否などを確認
4. プログレスバーを眺めながら変換完了を待つ

### 非対話モード(フラグ指定)

フラグが揃っているとそのまま実行されます。スクリプトやCIからも使えます。

```sh
# 単一ファイルをjpgとpngに変換
heic-converter IMG_0001.heic --format jpg,png

# フォルダ内のHEICを再帰的にすべてwebpへ、./convertedに出力
heic-converter ./photos --format webp --recursive --output ./converted

# 既存ファイルを上書き、JPEG品質を指定
heic-converter ./photos -f jpg -q 85 --overwrite
```

| フラグ | 短縮 | 説明 | デフォルト |
|---|---|---|---|
| `--format` | `-f` | 出力形式(カンマ区切りで複数指定可) | 対話で選択 |
| `--output` | `-o` | 出力先ディレクトリ | 元ファイルと同じ場所 |
| `--recursive` | `-r` | フォルダを再帰的に処理 | `false` |
| `--overwrite` | | 既存ファイルを上書き | `false`(既存ファイルはエラー) |
| `--quality` | `-q` | エンコード品質 1-100(jpg / webp) | `90` |

### 対応形式

| 出力形式 | 圧縮 | 備考 |
|---|---|---|
| jpg | 非可逆 | `--quality` 対応 |
| png | 可逆 | |
| webp | 非可逆 | `--quality` 対応 |
| tiff | 可逆 (deflate) | |
| bmp | 無圧縮 | |
| gif | 256色に減色 | |

### Webサーバーモード

`serve` サブコマンドでWebサーバーとして起動できます。単一エンドポイントで **gRPC / gRPC-Web / Connect (HTTP+JSON)** の3プロトコルを受け付けます。

```sh
heic-converter serve --port 8080
```

| フラグ | 説明 | デフォルト |
|---|---|---|
| `--host` | バインドするアドレス | `0.0.0.0` |
| `--port` | リッスンするポート | `8080` |
| `--max-request-bytes` | リクエストサイズ上限(バイト) | 64MiB |
| `--allowed-origins` | ブラウザからのアクセスを許可するCORSオリジン(カンマ区切り) | なし(CORS無効) |

```sh
# curlで変換(Connectプロトコル + JSON、画像はbase64)
curl -X POST http://localhost:8080/heic.v1.ConvertService/Convert \
  -H "Content-Type: application/json" \
  -d "{\"image\":\"$(base64 -i photo.heic)\",\"formats\":[\"jpg\",\"png\"],\"quality\":90}"

# 対応形式の一覧(副作用がないためGET可)
curl "http://localhost:8080/heic.v1.ConvertService/ListFormats?connect=v1&encoding=json&message=%7B%7D"

# gRPC(サーバーリフレクション対応なのでprotoファイル不要)
grpcurl -plaintext -d '{"image":"<base64>","formats":["jpg"]}' \
  localhost:8080 heic.v1.ConvertService/Convert

# ヘルスチェック(gRPC Health Checking Protocol準拠)
curl -X POST http://localhost:8080/grpc.health.v1.Health/Check \
  -H "Content-Type: application/json" -d '{}'
```

- 変換結果はレスポンスに直接含まれ、サーバー側には何も保存されません(ステートレス)
- TLS終端・認証はリバースプロキシ側の責務とし、サーバー自身は平文HTTP/2(h2c)で待ち受けます
- スキーマは [proto/heic/v1/convert.proto](proto/heic/v1/convert.proto) を参照してください

### Webフロントエンド

`web/` にブラウザ用のUI(Vite + React + TypeScript)があります。HEICを選んで形式を選ぶだけで変換でき、スマホではOSの共有シート経由でGoogle Drive保存やメール添付までそのまま行えます。変換画像はサーバーにもブラウザにも永続化されません。

```sh
# 1. APIサーバーを起動
heic-converter serve

# 2. フロント開発サーバーを起動(RPCはViteのプロキシ経由でserveへ届く)
cd web
npm install
npm run dev
```

開発サーバーはRPCパスを `localhost:8080` へプロキシするため、CORSや接続先の設定は不要です。別ホストのAPIサーバーへ直接つなぐ場合のみ `VITE_API_URL=<サーバーURL> npm run dev` で指定し、serve側で `--allowed-origins <フロントのオリジン>` を許可してください。

要件・設計は [doc/frontend-web-ui/prd.md](doc/frontend-web-ui/prd.md) を参照してください。

### Docker Composeで起動

GoもNodeもインストールせず、UI+APIをまとめてローカルで動かせます。

```sh
docker compose up --build
```

起動後 http://localhost:3000 を開くとWeb UIが使えます。

- `web`(nginx)が静的ファイルを配信し、RPCパスを `api` コンテナへリバースプロキシします(同一オリジンのためCORS不要)
- `api` はホストへポート公開していません(RPCも http://localhost:3000 に対して叩けます)
- 変換画像はどのコンテナにも保存されません(ステートレス)

## 開発

開発環境は [devcontainer](.devcontainer/devcontainer.json) で再現できます(Go 1.26.4 / Node 22)。

```sh
# Go(backend/): テスト・カバレッジ・ビルド
cd backend
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
CGO_ENABLED=0 go build ./cmd/heic-converter

# フロントエンド(web/): テスト・lint・ビルド
cd web && npm test && npm run lint && npm run build

# protoからGo/TypeScript両方のコードを生成(リポジトリルートで実行)
buf generate
```

### アーキテクチャ

クリーンアーキテクチャを採用しています。domain層がデコーダ/エンコーダ/ストレージのinterface(port)を定義し、infra層がそれを実装するため、ライブラリの差し替えが容易です。入力層(presentation)はusecaseのみに依存しており、将来CLI以外(HTTP APIなど)の入口を追加できます。

```
proto/heic/v1/          APIスキーマ (protobuf、Go/TS共通の源泉)
backend/                Go本体 (CLI + Webサーバー)
  cmd/heic-converter/   エントリポイント (DI)
  internal/
    domain/             モデルとport定義 (外部ライブラリ非依存)
    usecase/            変換ロジック (並列処理・fail-soft・進捗通知)
    infra/              portの実装 (HEICデコーダ・各エンコーダ・ローカルFS)
    gen/                bufによる生成コード (connect-go)
    presentation/cli/   CLI (cobra + charmbracelet)
    presentation/api/   Webサーバー (connect-rpc)
web/                    フロントエンド (Vite + React, atomic design)
  src/gen/              bufによる生成コード (protoc-gen-es)
  src/lib/ src/hooks/   RPC・ブラウザAPIの知識を隔離する層
  src/components/       atoms / molecules / organisms / templates / pages
```

HEICのデコードには [gen2brain/heic](https://github.com/gen2brain/heic)(デコーダをWASM化しpure Goランタイムで実行)を採用しており、これが「cgoなしの単一バイナリ」を実現しています。

アーキテクチャの詳細(層の依存ルール、入力層抽象化の解説)は [doc/architecture/clean-architecture-overview.md](doc/architecture/clean-architecture-overview.md) を、各機能の要件は [doc/cli-converter/prd.md](doc/cli-converter/prd.md) と [doc/connect-rpc-server/prd.md](doc/connect-rpc-server/prd.md) を参照してください。
