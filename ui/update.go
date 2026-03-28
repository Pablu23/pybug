package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.HandleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case StdoutMsg:
		m.messages = append(m.messages, string(msg))
		return m, ListenBridge(m.listenBridge)
	}

	return m, nil
}

func (m Model) HandleKeyMsg(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "k":
		m.cursor = clamp(0, m.height, m.cursor-1)
	case "j":
		m.cursor = clamp(0, min(m.height, m.lineCount-1), m.cursor+1)
	case "b":
		m.bridge.Breakpoint(m.currentFile, m.cursor)
		if file, ok := m.breakpoints[m.currentFile]; ok {
			m.breakpoints[m.currentFile] = append(file, m.cursor+1)
		} else {
			m.breakpoints[m.currentFile] = []int{
				m.cursor + 1,
			}
		}
	case "c":
		m.bridge.Continue()
	}

	return m, nil
}

func clamp(minimum, maximum, val int) int {
	val = max(minimum, val)
	val = min(maximum, val)
	return val
}
