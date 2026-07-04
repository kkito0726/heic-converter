# 処理シーケンス図(CLI)

heic-converter CLIの処理フローをMermaidのシーケンス図で示す。実装は
[実装方針](./implementation-plan.md)のクリーンアーキテクチャ構成
(`cmd` → `presentation/cli` → `usecase` → `domain/port` ← `infra`)に対応する。

## 1. 起動〜モード分岐

`main`から`cli.New(conv).ExecuteContext`が呼ばれた後、標準入出力がTTYかどうかで
対話モード(リッチUI)と非対話モード(プレーンテキスト)に分岐する
(`internal/presentation/cli/convert.go`の`run`。コマンド定義は1コマンド=1ファイルで、
root.goはコマンドツリーの組み立てのみを行う)。

```mermaid
sequenceDiagram
    actor User as ユーザー
    participant Main as cmd/main.go
    participant Root as cli.New().Run (convert.go)
    participant Interactive as interactive.go
    participant TUI as tui.go (bubbletea)
    participant Plain as runPlain (convert.go)
    participant Conv as usecase.Converter

    User->>Main: heic-converter [path] [flags]
    Main->>Root: cli.New(conv).ExecuteContext(ctx)
    Root->>Root: isTerminal() を判定

    alt TTYで実行(対話モード)
        Root->>Root: printLogo() でロゴを表示
        Root->>Interactive: runInteractive(conv, path, opts)
        Interactive->>User: askPath() パス入力(未指定時)
        Interactive->>Conv: SupportedFormats()
        Conv-->>Interactive: 対応形式一覧
        Interactive->>User: askFormats() チェックボックスで形式選択(未指定時)
        Interactive->>User: askOptions() 出力先/再帰/上書きを確認(未指定時)
        Interactive-->>Root: usecase.ConvertInput
        Root->>TUI: runWithTUI(ctx, conv, input, out)
        Note over TUI,Conv: 詳細は「2. 変換処理」を参照
        TUI-->>User: プログレスバー・結果を表示
    else 非TTY(パイプ・CI)
        Root->>Plain: runPlain(cmd, path, opts, conv)
        Plain->>Plain: buildInput() でフラグを検証・変換
        Plain->>Conv: Convert(ctx, input, textProgress)
        Note over Plain,Conv: 詳細は「2. 変換処理」を参照
        Plain-->>User: テキストで進捗・サマリを出力
    end
```

## 2. 変換処理(usecase.Converter.Convert)

対話・非対話いずれのモードでも、最終的に呼ばれる`Converter.Convert`の内部処理は
共通(`internal/usecase/convert.go`)。ProgressObserverの実装だけが
`teaObserver`(TUI)か`textProgress`(プレーン)かで切り替わる。

```mermaid
sequenceDiagram
    participant Caller as 呼び出し元 (tui.go / convert.go)
    participant Conv as usecase.Converter
    participant Storage as port.FileStorage (infra/storage.LocalFS)
    participant Decoder as port.ImageDecoder (infra/decoder.HEIC)
    participant Encoder as port.ImageEncoder (infra/encoder.*)
    participant Obs as ProgressObserver (teaObserver / textProgress)

    Caller->>Conv: Convert(ctx, ConvertInput, obs)
    Conv->>Conv: validateFormats() で形式を検証
    Conv->>Conv: FindSources(path, recursive)
    Conv->>Storage: FindFiles(path, recursive)
    Storage-->>Conv: ファイルパス一覧
    Conv->>Decoder: CanDecode(path) で.heic/.heifを絞り込み
    Conv->>Obs: OnStart(len(sources))

    par 各ソースファイルを並列変換(errgroup, 上限=NumCPU)
        Conv->>Conv: convertOne(src, input)
        Conv->>Storage: Open(src)
        Storage-->>Conv: io.ReadCloser
        Conv->>Decoder: Decode(r)
        Decoder-->>Conv: image.Image (1回だけデコード)
        loop 選択された各出力形式
            Conv->>Conv: OutputPath(src, outputDir, format)
            Conv->>Storage: Create(path, overwrite)
            alt overwrite=false かつ既存ファイルあり
                Storage-->>Conv: エラー (fs.ErrExist)
            else
                Storage-->>Conv: io.WriteCloser
                Conv->>Encoder: Encode(w, img, opts)
                Encoder-->>Conv: (成功 or エラー)
            end
        end
        Conv->>Obs: OnFileDone(result, done, total)
    end

    Conv-->>Caller: []ConversionResult (fail-soft: 失敗はresultに記録)
```

## ポイント

- **fail-soft**: 1ファイルの失敗(デコード失敗・既存ファイルなど)は`ConversionResult.Err`に記録されるだけで、他ファイルの変換や関数全体は止まらない
- **1回デコード・複数エンコード**: 複数形式を選択しても`Decode`は1回だけ実行し、結果を使い回して各`Encoder`に渡す
- **並列度**: `ConvertInput.Parallelism`が0以下の場合は`runtime.NumCPU()`が上限になる(`errgroup.SetLimit`)
- **層の依存方向**: `usecase.Converter`は`port.ImageDecoder` / `port.ImageEncoder` / `port.FileStorage`というinterfaceにのみ依存し、`infra`の具体実装(HEICデコーダやローカルFSなど)を知らない
