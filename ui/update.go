package ui

import (
	"log/slog"
	"strings"

	"git.pablu.de/pablu/pybug/ui/codeviewer"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updatedCv, cmd := m.codeViewer.Update(msg)
	m.codeViewer = updatedCv.(codeviewer.CodeViewer)

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
	case codeviewer.ExecutionStoppedMsg:
		return m, tea.Batch(ListenBridgeExecutionsStopped(m.listenBridgeExecutionsStopped), m.GetLocals())
	case LocalsMsg:
		m.currLocals = map[string]any(msg)
		m.localsViewer.SetContent(strings.Join(flattenDict(m.currLocals, 0), "\n"))
	}

	return m, cmd
}

func (m Model) GetLocals() tea.Cmd {
	return func() tea.Msg {
		vars, err := m.bridge.Locals()
		if err != nil {
			slog.Error("couldnt get vars", "error", err)
			return nil
		}

		return LocalsMsg(vars)
	}
}

func (m Model) UpdateWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	editorHeight := msg.Height * 70 / 100
	outputHeight := msg.Height - editorHeight - 4

	m.codeViewer.Width = msg.Width
	m.codeViewer.Height = editorHeight - 2

	m.stdoutOutput.Width = msg.Width / 2
	m.stdoutOutput.Height = outputHeight

	m.localsViewer.Width = msg.Width / 2
	m.localsViewer.Height = outputHeight

	m.stdoutOutput.SetContent(strings.Join(m.messages, ""))

	return m, nil
}

func (m Model) HandleKeyMsg(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "b":
		lineNumber := m.codeViewer.Cursor + 1

		set, err := m.bridge.Breakpoint(m.currentFile, lineNumber)
		if err != nil {
			slog.Error("could not set or unset breakpoint", "error", err)
		}
		if file, ok := m.breakpoints[m.currentFile]; ok {
			m.breakpoints[m.currentFile] = append(file, lineNumber)
		} else {
			m.breakpoints[m.currentFile] = []int{
				lineNumber,
			}
		}
		return m, func() tea.Msg {
			// check if this is in currently viewed file
			return codeviewer.BreakpointMsg{
				Added: set,
				Line:  lineNumber,
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
