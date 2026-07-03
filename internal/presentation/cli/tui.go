package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kkito0726/heic-converter/internal/domain/model"
	"github.com/kkito0726/heic-converter/internal/usecase"
)

// ErrCanceledは実行中の変換をユーザーが中断したときに返される。
var ErrCanceled = errors.New("conversion canceled")

// extraProgramOptionsは進捗UIをヘッドレスで(レンダラーなし・TTY入力なしで)
// テストできるようにするためのもの。
var extraProgramOptions []tea.ProgramOption

type startedMsg struct{ total int }

type fileDoneMsg struct {
	res         model.ConversionResult
	done, total int
}

type finishedMsg struct {
	results []model.ConversionResult
	err     error
}

// teaObserverはusecaseの進捗コールバックをbubbletea用のメッセージに橋渡しする。
type teaObserver struct{ p *tea.Program }

var _ usecase.ProgressObserver = (*teaObserver)(nil)

func (o *teaObserver) OnStart(total int) { o.p.Send(startedMsg{total: total}) }
func (o *teaObserver) OnFileDone(res model.ConversionResult, done, total int) {
	o.p.Send(fileDoneMsg{res: res, done: done, total: total})
}

// maxVisibleResultsは変換中に画面上に残すファイルごとの行数を制限する。
// 完全なサマリは最後にまとめて出力される。
const maxVisibleResults = 8

type runModel struct {
	spin     spinner.Model
	bar      progress.Model
	total    int
	done     int
	failed   int
	recent   []string
	results  []model.ConversionResult
	err      error
	finished bool
	cancel   context.CancelFunc
}

var _ tea.Model = runModel{}

func newRunModel(cancel context.CancelFunc) runModel {
	return runModel{
		spin:   spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(styleSpinner)),
		bar:    progress.New(progress.WithDefaultGradient()),
		cancel: cancel,
	}
}

func (m runModel) Init() tea.Cmd { return m.spin.Tick }

func (m runModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.cancel()
			return m, tea.Quit
		}
		return m, nil
	case startedMsg:
		m.total = msg.total
		return m, nil
	case fileDoneMsg:
		m.done = msg.done
		m.total = msg.total
		if !msg.res.Succeeded() {
			m.failed++
		}
		m.recent = appendCapped(m.recent, resultLine(msg.res), maxVisibleResults)
		return m, nil
	case finishedMsg:
		m.results = msg.results
		m.err = msg.err
		m.finished = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m runModel) View() string {
	if m.finished {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s Converting %d/%d\n", m.spin.View(), m.done, m.total))
	percent := 0.0
	if m.total > 0 {
		percent = float64(m.done) / float64(m.total)
	}
	b.WriteString(m.bar.ViewAs(percent))
	b.WriteString("\n\n")
	for _, line := range m.recent {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteString(styleMuted.Render("press q to cancel"))
	b.WriteByte('\n')
	return b.String()
}

func appendCapped(lines []string, line string, capacity int) []string {
	lines = append(lines, line)
	if len(lines) > capacity {
		return lines[len(lines)-capacity:]
	}
	return lines
}

func resultLine(res model.ConversionResult) string {
	name := filepath.Base(res.SourcePath)
	if res.Succeeded() {
		return fmt.Sprintf("%s %s", styleSuccess.Render("✓"), name)
	}
	return fmt.Sprintf("%s %s", styleError.Render("✗"), name)
}

// runWithTUIはライブの進捗画面を描画しながら変換を実行する。
func runWithTUI(ctx context.Context, conv *usecase.Converter, in usecase.ConvertInput, out io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := append([]tea.ProgramOption{tea.WithOutput(out)}, extraProgramOptions...)
	p := tea.NewProgram(newRunModel(cancel), opts...)
	go func() {
		results, err := conv.Convert(ctx, in, &teaObserver{p: p})
		p.Send(finishedMsg{results: results, err: err})
	}()

	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("progress ui failed: %w", err)
	}
	m, ok := final.(runModel)
	if !ok || !m.finished {
		return ErrCanceled
	}
	if m.err != nil {
		return m.err
	}
	return printStyledSummary(out, m.results)
}

// printStyledSummaryは最終レポートを出力し、1件でも失敗していればエラーを
// 返す。これによりプロセスは非ゼロで終了する。
func printStyledSummary(w io.Writer, results []model.ConversionResult) error {
	var failed []model.ConversionResult
	for _, r := range results {
		if !r.Succeeded() {
			failed = append(failed, r)
		}
	}
	fmt.Fprintf(w, "\n%s %d succeeded  %s %d failed\n",
		styleSuccess.Render("✓"), len(results)-len(failed),
		styleError.Render("✗"), len(failed))
	for _, r := range failed {
		fmt.Fprintf(w, "  %s %s: %v\n", styleError.Render("✗"), r.SourcePath, r.Err)
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d of %d file(s) failed", len(failed), len(results))
	}
	return nil
}
