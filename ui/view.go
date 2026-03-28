package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var panelStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)

func flattenDict(m map[string]interface{}, indent int) []string {
	lines := []string{}
	prefix := strings.Repeat("  ", indent)
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			lines = append(lines, fmt.Sprintf("%s%s:", prefix, k))
			lines = append(lines, flattenDict(val, indent+1)...)
		default:
			lines = append(lines, fmt.Sprintf("%s%s: %v", prefix, k, val))
		}
	}
	return lines
}

func (m Model) View() string {
	hFrame, wFrame := panelStyle.GetFrameSize()
	topPanel := panelStyle.
		Height(m.codeViewer.Height - hFrame).
		Width(m.codeViewer.Width - wFrame).
		Render(m.codeViewer.View())

	bottomLeftPanel := panelStyle.Height(m.stdoutOutput.Height - hFrame).
		Width(m.stdoutOutput.Width - wFrame).
		Render(m.stdoutOutput.View())

	bottomRightPanel := panelStyle.Height(m.stdoutOutput.Height - hFrame).
		Width(m.stdoutOutput.Width - wFrame).
		Render(m.localsViewer.View())

	bottomPanel := lipgloss.JoinHorizontal(lipgloss.Top, bottomLeftPanel, bottomRightPanel)

	return lipgloss.JoinVertical(lipgloss.Top, topPanel, bottomPanel)
}
