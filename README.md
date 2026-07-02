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

### Goでインストール

```sh
go install github.com/kkito0726/heic-converter/cmd/heic-converter@latest
```

### ソースからビルド

```sh
git clone https://github.com/kkito0726/heic-converter.git
cd heic-converter
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

## 開発

開発環境は [devcontainer](.devcontainer/devcontainer.json) で再現できます(Go 1.26.4)。

```sh
# テスト
CGO_ENABLED=0 go test ./...

# カバレッジ
CGO_ENABLED=0 go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# ビルド
CGO_ENABLED=0 go build ./cmd/heic-converter
```

### アーキテクチャ

クリーンアーキテクチャを採用しています。domain層がデコーダ/エンコーダ/ストレージのinterface(port)を定義し、infra層がそれを実装するため、ライブラリの差し替えが容易です。入力層(presentation)はusecaseのみに依存しており、将来CLI以外(HTTP APIなど)の入口を追加できます。

```
cmd/heic-converter/   エントリポイント (DI)
internal/
  domain/             モデルとport定義 (外部ライブラリ非依存)
  usecase/            変換ロジック (並列処理・fail-soft・進捗通知)
  infra/              portの実装 (HEICデコーダ・各エンコーダ・ローカルFS)
  presentation/cli/   CLI (cobra + charmbracelet)
```

HEICのデコードには [gen2brain/heic](https://github.com/gen2brain/heic)(デコーダをWASM化しpure Goランタイムで実行)を採用しており、これが「cgoなしの単一バイナリ」を実現しています。

設計の詳細は [doc/prd/PRD.md](doc/prd/PRD.md) と [doc/prd/IMPLEMENTATION.md](doc/prd/IMPLEMENTATION.md) を参照してください。
