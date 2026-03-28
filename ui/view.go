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

var panelStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())

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

	frameW, frameH := panelStyle.GetFrameSize()
	topHeight := m.height * 70 / 100

	code := panelStyle.Width(m.width - frameW).Height(topHeight - frameH).Render(out.String())
	m.viewport.Height = m.height - (topHeight + frameH)
	m.viewport.Width = m.width - frameW
	m.viewport.SetContent(strings.Join(m.messages, ""))
	m.viewport.GotoBottom()
	output := panelStyle.Width(m.viewport.Width).Height(m.viewport.Height).Render(m.viewport.View())

	return lipgloss.JoinVertical(lipgloss.Top, code, output)
}
