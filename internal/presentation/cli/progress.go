package cli

import (
	"fmt"
	"io"

	"github.com/kkito0726/heic-converter/internal/domain/model"
)

// textProgress is a plain-text ProgressObserver used in non-interactive runs.
type textProgress struct {
	out io.Writer
}

func newTextProgress(out io.Writer) *textProgress {
	return &textProgress{out: out}
}

// OnStart implements usecase.ProgressObserver.
func (p *textProgress) OnStart(total int) {
	fmt.Fprintf(p.out, "Converting %d file(s)...\n", total)
}

// OnFileDone implements usecase.ProgressObserver.
func (p *textProgress) OnFileDone(res model.ConversionResult, done, total int) {
	if res.Succeeded() {
		fmt.Fprintf(p.out, "[%d/%d] OK   %s -> %v\n", done, total, res.SourcePath, res.Outputs)
		return
	}
	fmt.Fprintf(p.out, "[%d/%d] FAIL %s: %v\n", done, total, res.SourcePath, res.Err)
}

// printSummary writes the final success/failure counts and returns the number
// of failed files.
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
