# 実装方針: heic-converter

[PRD](./PRD.md) を実現するための技術選定・アーキテクチャ・実装計画。

## 1. 技術選定

### 1.1 HEIC デコード(最重要の選定)

「単一バイナリ・ユーザー環境に依存しない」が最優先要件のため、**cgo を使わない**ことが絶対条件。

| 候補 | 方式 | 評価 |
|---|---|---|
| **`github.com/gen2brain/heic`(採用)** | libheif を WASM にコンパイルし、pure Go の WASM ランタイム [wazero](https://wazero.io) で実行 | ✅ cgo 不要・`CGO_ENABLED=0` でビルド可・クロスコンパイル容易・`image.Image` を返す標準的な API |
| `github.com/adrium/goheif` (jdeng/goheif 系) | libde265 を cgo でバンドル | ❌ cgo 必須。クロスコンパイルに各OSのCツールチェーンが必要 |
| `github.com/strukturag/libheif` の Go binding | システムの libheif に動的リンク | ❌ ユーザー環境に libheif のインストールが必要 |
| macOS `sips` 等の外部コマンド呼び出し | exec | ❌ OS 依存。Windows/Linux で動かない |

- wazero 方式は純正 C 実装よりデコードは遅いが、並列化で補える(NFR の「実用的な時間」は満たせる)
- domain 層にデコーダの interface を切るため、将来より良いライブラリが出れば infra 層の差し替えだけで移行できる

### 1.2 エンコード(出力形式)

| 形式 | ライブラリ | 備考 |
|---|---|---|
| jpg | `image/jpeg`(標準) | quality オプション対応 |
| png | `image/png`(標準) | |
| gif | `image/gif`(標準) | |
| tiff | `golang.org/x/image/tiff` | 準標準 |
| bmp | `golang.org/x/image/bmp` | 準標準 |
| webp | `github.com/gen2brain/webp` | 標準/x/image には encoder がないため。heic と同じ wazero 方式で cgo 不要 |

### 1.3 CLI / UI

すべて [charmbracelet](https://charm.sh) エコシステムで統一する(pure Go・実績・デザイン一貫性)。

| 用途 | ライブラリ |
|---|---|
| 対話フォーム(チェックボックス複数選択・テキスト入力・確認) | `charmbracelet/huh` |
| プログレスバー・スピナー・実行中画面 | `charmbracelet/bubbletea` + `charmbracelet/bubbles` |
| 配色・スタイリング | `charmbracelet/lipgloss` |
| フラグパース | `spf13/cobra`(将来の `serve` サブコマンド追加が容易) |
| アスキーアートロゴ | 生成済みのものを Go ソースに埋め込み、lipgloss でグラデーション着色(実行時生成ライブラリは不要) |

### 1.4 その他

- EXIF Orientation の読み取り: `github.com/rwcarlsen/goexif` または gen2brain/heic が返す情報を利用し、デコード後に回転補正をかける
- 並列処理: `golang.org/x/sync/errgroup` + `SetLimit(runtime.NumCPU())` の worker 制御

## 2. アーキテクチャ

クリーンアーキテクチャ。**依存の向きは常に内側(domain)へ**。domain 層が interface を定義し、infra 層が実装する。

```
cmd/
  heic-converter/
    main.go              # エントリポイント。DI(依存の組み立て)のみ行う
internal/
  domain/                # 中心。外部ライブラリに依存しない
    model/
      image.go           # Image, Format などのエンティティ・値オブジェクト
      task.go            # ConversionTask, ConversionResult
    port/                # infra が実装すべき interface(出力ポート)
      decoder.go         # ImageDecoder interface
      encoder.go         # ImageEncoder interface
      storage.go         # FileStorage interface(探索・読み書き)
  usecase/
    convert.go           # 変換ユースケース(単一/複数、並列制御、進捗通知)
    progress.go          # 進捗を通知する Observer interface(UI層が実装)
  infra/                 # port の実装。外部ライブラリはここに閉じ込める
    decoder/
      heic.go            # gen2brain/heic による ImageDecoder 実装
    encoder/
      jpeg.go png.go webp.go tiff.go bmp.go gif.go
      registry.go        # Format → Encoder の解決
    storage/
      localfs.go         # ローカルファイルシステム実装
  presentation/          # 入力層。usecase だけに依存
    cli/
      root.go            # cobra コマンド定義・フラグ
      interactive.go     # huh による対話フロー
      runner.go          # bubbletea 進捗画面(usecase の進捗 Observer を実装)
      theme.go           # lipgloss カラーテーマ・ロゴ
    api/                 # (将来)HTTP サーバー実装を置く場所
```

### 設計の要点

- **デコーダ/エンコーダ差し替え**: `domain/port` の interface を infra が実装。wazero 版が遅い・不具合となった場合も infra の1ファイル差し替えで済む
- **入力層の抽象化**: presentation は usecase の公開メソッドのみを呼ぶ。CLI 固有の型(bubbletea の Msg 等)を usecase に持ち込まない。将来 `presentation/api` を追加しても usecase は無変更
- **進捗通知**: usecase → UI 方向は `ProgressObserver` interface(またはイベント channel)で通知し、usecase が UI 実装を知らない状態を保つ

### 主要 interface(イメージ)

```go
// domain/port/decoder.go
type ImageDecoder interface {
    Decode(r io.Reader) (image.Image, error)
    CanDecode(path string) bool
}

// domain/port/encoder.go
type ImageEncoder interface {
    Encode(w io.Writer, img image.Image, opts model.EncodeOptions) error
    Format() model.Format
}

// domain/port/storage.go
type FileStorage interface {
    FindImages(path string, recursive bool) ([]string, error)
    Open(path string) (io.ReadCloser, error)
    Create(path string, overwrite bool) (io.WriteCloser, error)
}
```

## 3. 処理フロー

```
main → cobra(フラグ解釈)
  ├─ フラグが揃っている → 非対話でそのまま usecase.Convert 実行
  └─ 不足している      → huh で対話(パス入力 → 形式チェックボックス → オプション確認)
                          ↓
        usecase.Convert(tasks, observer)
          ├─ storage.FindImages でターゲット列挙(ファイル or フォルダを判定)
          ├─ errgroup で並列変換(decode 1回 → 選択された全形式に encode)
          └─ 進捗を observer へ通知 → bubbletea がプログレスバー描画
                          ↓
        結果サマリ表示(成功/失敗件数・失敗理由)
```

- 1ファイルの失敗は `ConversionResult` にエラーとして記録して続行(fail-soft)
- 複数形式選択時はデコード結果を使い回し、デコードは1ファイル1回に抑える

## 4. 開発環境(devcontainer)

- ベース: `mcr.microsoft.com/devcontainers/go`(Go 1.26)
- `CGO_ENABLED=0` を環境変数でデフォルト化し、cgo 依存の混入をビルド時に検知
- ツール: `golangci-lint`, `gotestsum`
- `postCreateCommand` で `go mod download`

## 5. テスト方針(カバレッジ 80% 以上・TDD)

| レイヤ | 方針 |
|---|---|
| domain / usecase | port の interface をモックした純粋なユニットテスト。並列変換・fail-soft・進捗通知を重点的に |
| infra/decoder | `testdata/` に小さな実 HEIC ファイルを置き、実デコードのゴールデンテスト(サイズ・向きの検証) |
| infra/encoder | `image.Image` → エンコード → 標準ライブラリで再デコードして round-trip 検証 |
| presentation/cli | 対話フローはロジック(選択結果 → usecase 引数への変換)を関数に切り出して単体テスト。E2E は非対話モードをバイナリ実行で検証 |

- 各機能とも Red → Green → Refactor の TDD で進める

## 6. 実装フェーズ

### Phase 1: コア変換(MVP)
1. リポジトリ初期化・devcontainer・CI(lint + test)
2. domain のモデルと port 定義
3. infra: heic デコーダ + jpg/png エンコーダ + localfs
4. usecase: 単一ファイル変換
5. 最小 CLI(非対話・フラグのみ)で end-to-end 動作確認

### Phase 2: 一括変換
6. フォルダ探索(再帰オプション)+ errgroup 並列変換
7. fail-soft と結果サマリ
8. EXIF Orientation 補正

### Phase 3: 対話 UI・リッチ化
9. huh による対話フロー(パス入力・形式チェックボックス・確認)
10. bubbletea + bubbles のプログレス画面、lipgloss テーマ、アスキーアートロゴ

### Phase 4: 形式拡充・配布
11. webp / tiff / bmp / gif エンコーダ追加
12. goreleaser でクロスビルド(darwin/linux/windows × amd64/arm64)・リリース自動化

## 7. リスクと対応

| リスク | 対応 |
|---|---|
| wazero 方式のデコードが遅い | 並列化で緩和。port 抽象化済みなので必要なら cgo 版ビルドを別タグで提供する余地も残る |
| HEIC の亜種(連写・深度情報付き等)でデコード失敗 | fail-soft で全体は止めない。testdata に多様なサンプルを追加して検証 |
| Windows ターミナルでの描画崩れ | charmbracelet 系は Windows 対応済み。CI で最低限の起動確認を行う |
| 巨大フォルダでのメモリ消費 | worker 数を CPU コア数で制限し、同時にメモリへ載る画像数を抑制 |
