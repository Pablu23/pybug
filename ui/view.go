package ui

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

var panelStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)

func (m Model) View() string {
	var buf bytes.Buffer
	lexer := lexers.Get("python")
	style := styles.Get("monokai")
	formatter := formatters.Get("terminal16m")

	iterator, _ := lexer.Tokenise(nil, m.text)
	formatter.Format(&buf, style, iterator)

	var out strings.Builder

	lines := strings.Split(buf.String(), "\n")
	breakpoints := m.breakpoints[m.currentFile]
	for i, line := range lines {
		breakpoint := " "
		cursor := " "

		if slices.Contains(breakpoints, i+1) {
			breakpoint = "O"
		}

		if i == m.cursor {
			cursor = ">"
		}
		fmt.Fprintf(&out, "%-4d%s%s %s\n", i+1, breakpoint, cursor, line)
	}

	m.codeViewer.SetContent(out.String())

	hFrame, wFrame := panelStyle.GetFrameSize()
	topPanel := panelStyle.
		Height(m.codeViewer.Height - hFrame).
		Width(m.codeViewer.Width - wFrame).
		Render(m.codeViewer.View())

	bottomPanel := panelStyle.Height(m.stdoutOutput.Height - hFrame).
		Width(m.stdoutOutput.Width - wFrame).
		Render(m.stdoutOutput.View())

	return lipgloss.JoinVertical(lipgloss.Top, topPanel, bottomPanel)
}
