package ui

import (
	"strings"

	"git.pablu.de/pablu/pybug/internal/bridge"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	currentFile string
	cursor      int
	text        string
	width       int
	height      int
	lineCount   int
	bridge      *bridge.Bridge

	listenBridge chan string

	messages []string
	viewport viewport.Model

	breakpoints map[string][]int
}

func NewModel(b *bridge.Bridge, file string, text string) Model {
	c := b.Subscribe()

	vp := viewport.New(0, 0)

	return Model{
		currentFile:  file,
		text:         text,
		cursor:       0,
		lineCount:    len(strings.Split(text, "\n")),
		bridge:       b,
		breakpoints:  make(map[string][]int),
		listenBridge: c,
		messages:     make([]string, 0),
		viewport:     vp,
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
