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

// ErrCanceled is returned when the user aborts a running conversion.
var ErrCanceled = errors.New("conversion canceled")

// extraProgramOptions lets tests run the progress UI headless
// (no renderer, no TTY input).
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

// teaObserver bridges usecase progress callbacks into bubbletea messages.
type teaObserver struct{ p *tea.Program }

func (o *teaObserver) OnStart(total int) { o.p.Send(startedMsg{total: total}) }
func (o *teaObserver) OnFileDone(res model.ConversionResult, done, total int) {
	o.p.Send(fileDoneMsg{res: res, done: done, total: total})
}

// maxVisibleResults limits how many per-file lines stay on screen while
// converting; the full summary is printed at the end.
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

// runWithTUI executes the conversion while rendering a live progress screen.
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

// printStyledSummary prints the final report and returns an error when any
// file failed, so the process exits non-zero.
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
