// cliパッケージはコマンドライン向けのpresentation層。usecase層にのみ
// 依存するため、将来API向けのpresentationを追加しても同じコアを再利用できる。
package cli

import (
	"github.com/spf13/cobra"

	"github.com/kkito0726/heic-converter/backend/internal/usecase"
)

// Newは指定されたconverterを組み込んだルートコマンドを構築する。
// 各コマンドの定義は1コマンド=1ファイル(convert.go / serve.go)に置き、
// ここではコマンドツリーの組み立てのみを行う。
func New(conv *usecase.Converter) *cobra.Command {
	cmd := newConvertCmd(conv)
	cmd.AddCommand(newServeCmd(conv))
	return cmd
}
