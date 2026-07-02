package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorAccent  = lipgloss.Color("#EE6FF8")
	colorSuccess = lipgloss.Color("#04B575")
	colorError   = lipgloss.Color("#FF5F87")
	colorMuted   = lipgloss.Color("241")

	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleSpinner = lipgloss.NewStyle().Foreground(colorAccent)
)

var logoLines = []string{
	"██╗  ██╗███████╗██╗ ██████╗",
	"██║  ██║██╔════╝██║██╔════╝",
	"███████║█████╗  ██║██║     ",
	"██╔══██║██╔══╝  ██║██║     ",
	"██║  ██║███████╗██║╚██████╗",
	"╚═╝  ╚═╝╚══════╝╚═╝ ╚═════╝ CONVERTER",
}

// logoGradient colors the logo line by line from purple to pink.
var logoGradient = []string{
	"#7D56F4", "#945CF5", "#AB62F6", "#C267F7", "#D86BF8", "#EE6FF8",
}

func logo() string {
	var b strings.Builder
	for i, line := range logoLines {
		color := logoGradient[i%len(logoGradient)]
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		b.WriteString(style.Render(line))
		b.WriteByte('\n')
	}
	b.WriteString(styleMuted.Render("Convert HEIC/HEIF images to jpg, png, webp and more"))
	b.WriteByte('\n')
	return b.String()
}

func printLogo(w io.Writer) {
	fmt.Fprintln(w, logo())
}
