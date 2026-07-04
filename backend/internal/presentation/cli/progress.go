package cli

import (
	"fmt"
	"io"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
	"github.com/kkito0726/heic-converter/backend/internal/usecase"
)

// textProgressは非対話実行で使うプレーンテキストのProgressObserver。
type textProgress struct {
	out io.Writer
}

var _ usecase.ProgressObserver = (*textProgress)(nil)

func newTextProgress(out io.Writer) *textProgress {
	return &textProgress{out: out}
}

// OnStartはusecase.ProgressObserverを実装する。
func (p *textProgress) OnStart(total int) {
	fmt.Fprintf(p.out, "Converting %d file(s)...\n", total)
}

// OnFileDoneはusecase.ProgressObserverを実装する。
func (p *textProgress) OnFileDone(res model.ConversionResult, done, total int) {
	if res.Succeeded() {
		fmt.Fprintf(p.out, "[%d/%d] OK   %s -> %v\n", done, total, res.SourcePath, res.Outputs)
		return
	}
	fmt.Fprintf(p.out, "[%d/%d] FAIL %s: %v\n", done, total, res.SourcePath, res.Err)
}

// printSummaryは最終的な成功/失敗件数を出力し、失敗したファイル数を返す。
func printSummary(out io.Writer, results []model.ConversionResult) int {
	var failed int
	for _, r := range results {
		if !r.Succeeded() {
			failed++
		}
	}
	fmt.Fprintf(out, "\nDone: %d succeeded, %d failed\n", len(results)-failed, failed)
	return failed
}
