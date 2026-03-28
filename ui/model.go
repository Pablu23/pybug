package ui

import (
	"strings"

	"git.pablu.de/pablu/pybug/internal/bridge"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	currentFile string
	text        string
	textLines   int

	width  int
	height int

	bridge                        *bridge.Bridge
	listenBridge                  chan string
	listenBridgeExecutionsStopped chan bridge.ExecutionPoint

	currLocals         map[string]any
	currExecutionPoint bridge.ExecutionPoint

	messages     []string
	stdoutOutput viewport.Model
	codeViewer   viewport.Model
	localsViewer viewport.Model
	cursor       int

	breakpoints map[string][]int
}

func NewModel(b *bridge.Bridge, file string, text string) Model {
	c := b.Subscribe()
	c2 := b.SubscribeStopped()

	stdoutOutput := viewport.New(0, 0)
	codeViewer := viewport.New(0, 0)
	localsViewer := viewport.New(0, 0)

	return Model{
		currentFile: file,
		text:        text,
		textLines:   len(strings.Split(text, "\n")),

		bridge:                        b,
		listenBridge:                  c,
		listenBridgeExecutionsStopped: c2,

		currLocals: make(map[string]any),
		currExecutionPoint: bridge.ExecutionPoint{
			File: file,
			Line: 0,
		},
		breakpoints: make(map[string][]int),
		messages:    make([]string, 0),

		cursor:       0,
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
		return ExecutionStoppedMsg(msg)
	}
}

type LocalsMsg map[string]any
type ExecutionStoppedMsg bridge.ExecutionPoint
type StdoutMsg string

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		ListenBridge(m.listenBridge),
		ListenBridgeExecutionsStopped(m.listenBridgeExecutionsStopped),
	)
}
