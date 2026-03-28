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

	bridge       *bridge.Bridge
	listenBridge chan string

	messages     []string
	stdoutOutput viewport.Model
	codeViewer   viewport.Model
	cursor       int

	breakpoints map[string][]int
}

func NewModel(b *bridge.Bridge, file string, text string) Model {
	c := b.Subscribe()

	stdoutOutput := viewport.New(0, 0)
	codeViewer := viewport.New(0, 0)

	return Model{
		currentFile: file,
		text:        text,
		textLines:   len(strings.Split(text, "\n")),

		bridge:       b,
		listenBridge: c,

		breakpoints: make(map[string][]int),
		messages:    make([]string, 0),

		cursor:       0,
		codeViewer:   codeViewer,
		stdoutOutput: stdoutOutput,
	}
}

func ListenBridge(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch
		return StdoutMsg(msg)
	}
}

type StdoutMsg string

func (m Model) Init() tea.Cmd {
	return ListenBridge(m.listenBridge)
}
