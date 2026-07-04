package cli

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kkito0726/heic-converter/backend/internal/domain/model"
)

func TestRunModelProgress(t *testing.T) {
	m := newRunModel(func() {})

	next, _ := m.Update(startedMsg{total: 3})
	next, _ = next.Update(fileDoneMsg{
		res:  model.ConversionResult{SourcePath: "/p/a.heic", Outputs: []string{"/p/a.jpg"}},
		done: 1, total: 3,
	})
	next, _ = next.Update(fileDoneMsg{
		res:  model.ConversionResult{SourcePath: "/p/b.heic", Err: errors.New("boom")},
		done: 2, total: 3,
	})
	rm := next.(runModel)
	if rm.done != 2 || rm.total != 3 || rm.failed != 1 {
		t.Errorf("done=%d total=%d failed=%d, want 2/3/1", rm.done, rm.total, rm.failed)
	}
	view := rm.View()
	if !strings.Contains(view, "2/3") {
		t.Errorf("view missing progress count: %q", view)
	}
	if !strings.Contains(view, "a.heic") || !strings.Contains(view, "b.heic") {
		t.Errorf("view missing file lines: %q", view)
	}
}

func TestRunModelFinish(t *testing.T) {
	m := newRunModel(func() {})
	results := []model.ConversionResult{{SourcePath: "a.heic"}}
	next, cmd := m.Update(finishedMsg{results: results})
	rm := next.(runModel)
	if !rm.finished {
		t.Error("model should be finished")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if rm.View() != "" {
		t.Errorf("finished view should be empty, got %q", rm.View())
	}
}

func TestRunModelCancelKey(t *testing.T) {
	canceled := false
	m := newRunModel(func() { canceled = true })
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !canceled {
		t.Error("q should cancel the conversion context")
	}
	if cmd == nil {
		t.Error("q should quit the program")
	}
}

func TestAppendCapped(t *testing.T) {
	var lines []string
	for i := 0; i < 12; i++ {
		lines = appendCapped(lines, "x", 8)
	}
	if len(lines) != 8 {
		t.Errorf("len = %d, want 8", len(lines))
	}
}

func TestPrintStyledSummary(t *testing.T) {
	var sb strings.Builder
	err := printStyledSummary(&sb, []model.ConversionResult{
		{SourcePath: "a.heic", Outputs: []string{"a.jpg"}},
		{SourcePath: "b.heic", Err: errors.New("boom")},
	})
	if err == nil {
		t.Error("expected error when a file failed")
	}
	if !strings.Contains(sb.String(), "b.heic") {
		t.Errorf("summary should list failed file: %q", sb.String())
	}

	sb.Reset()
	if err := printStyledSummary(&sb, []model.ConversionResult{{SourcePath: "a.heic"}}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLogoRenders(t *testing.T) {
	got := logo()
	if !strings.Contains(got, "CONVERTER") {
		t.Errorf("logo missing tagline: %q", got)
	}
}
