package ui

import (
	"os"

	"git.pablu.de/pablu/pybug/internal/bridge"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(bridge *bridge.Bridge) error {
	buf, err := os.ReadFile("test.py")
	if err != nil {
		return err
	}

	m := NewModel(bridge, "test.py", string(buf))

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
