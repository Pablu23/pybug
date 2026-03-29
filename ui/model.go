package ui

import (
	"git.pablu.de/pablu/pybug/internal/bridge"
	"git.pablu.de/pablu/pybug/ui/codeviewer"
	"git.pablu.de/pablu/pybug/ui/variablesviewer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	currentFile string

	width  int
	height int

	bridge                        *bridge.Bridge
	listenBridge                  chan string
	listenBridgeExecutionsStopped chan bridge.ExecutionPoint

	currLocals map[string]any

	messages     []string
	stdoutOutput viewport.Model
	codeViewer   codeviewer.CodeViewer
	localsViewer variablesviewer.VariableViewer

	breakpoints map[string][]int
}

func NewModel(b *bridge.Bridge, file string, text string) Model {
	c := b.Subscribe()
	c2 := b.SubscribeStopped()

	stdoutOutput := viewport.New(0, 0)
	codeViewer := codeviewer.NewCodeViewer(text)
	localsViewer := variablesviewer.NewVariableViewer(map[string]any{})

	return Model{
		currentFile: file,

		bridge:                        b,
		listenBridge:                  c,
		listenBridgeExecutionsStopped: c2,

		currLocals:  make(map[string]any),
		breakpoints: make(map[string][]int),
		messages:    make([]string, 0),

		codeViewer:   codeViewer,
		stdoutOutput: stdoutOutput,
		localsViewer: localsViewer,
	}
}

func ListenBridge(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch
		return StdoutMsg(msg)
	}
}

func ListenBridgeExecutionsStopped(ch <-chan bridge.ExecutionPoint) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch
		return codeviewer.ExecutionStoppedMsg{
			Line: msg.Line,
			// TODO, WE IGNORE FILE HERE
		}
	}
}

type LocalsMsg map[string]any
type StdoutMsg string

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		ListenBridge(m.listenBridge),
		ListenBridgeExecutionsStopped(m.listenBridgeExecutionsStopped),
	)
}
