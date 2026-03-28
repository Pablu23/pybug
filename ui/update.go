package ui

import (
	"log/slog"
	"strings"

	"git.pablu.de/pablu/pybug/internal/bridge"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.HandleKeyMsg(msg)
	case tea.WindowSizeMsg:
		return m.UpdateWindowSize(msg)
	case StdoutMsg:
		m.messages = append(m.messages, string(msg))
		m.stdoutOutput.SetContent(strings.Join(m.messages, ""))
		m.stdoutOutput.GotoBottom()
		return m, ListenBridge(m.listenBridge)
	case ExecutionStoppedMsg:
		m.currExecutionPoint = bridge.ExecutionPoint(msg)
		return m, ListenBridgeExecutionsStopped(m.listenBridgeExecutionsStopped)
	}

	return m, nil
}

func (m Model) UpdateWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	editorHeight := msg.Height * 70 / 100
	outputHeight := msg.Height - editorHeight - 4

	m.codeViewer.Width = msg.Width
	m.codeViewer.Height = editorHeight

	m.stdoutOutput.Width = msg.Width
	m.stdoutOutput.Height = outputHeight

	m.stdoutOutput.SetContent(strings.Join(m.messages, ""))

	return m, nil
}

func (m Model) HandleKeyMsg(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "k":
		m.codeViewer.ScrollUp(1)
		m.cursor = max(0, m.cursor-1)
	case "j":
		m.codeViewer.ScrollUp(1)
		m.cursor = min(m.textLines-1, m.cursor+1)
	case "b":
		lineNumber := m.cursor + 1

		m.bridge.Breakpoint(m.currentFile, lineNumber)
		if file, ok := m.breakpoints[m.currentFile]; ok {
			m.breakpoints[m.currentFile] = append(file, lineNumber)
		} else {
			m.breakpoints[m.currentFile] = []int{
				lineNumber,
			}
		}
	case "s":
		m.bridge.Step()
	case "c":
		m.bridge.Continue()
	case "r":
		m.messages = make([]string, 0)
		m.stdoutOutput.SetContent("")
		err := m.bridge.Start()
		if err != nil {
			slog.Error("could not start brige", "error", err)
		}
	}

	return m, nil
}
