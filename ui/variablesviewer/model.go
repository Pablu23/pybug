package variablesviewer

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type VariableViewer struct {
	variables map[string]any

	Width, Height int
}

func NewVariableViewer(variables map[string]any) VariableViewer {
	return VariableViewer{
		variables: variables,
		Width:     0,
		Height:    0,
	}
}

func (v *VariableViewer) SetNewVariables(variables map[string]any) {
	v.variables = variables
}

func (v VariableViewer) Init() tea.Cmd {
	return nil
}

func (v VariableViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

func (v VariableViewer) View() string {
	out := strings.Builder{}

	keys := make([]string, 0, len(v.variables))
	for k := range v.variables {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	// TODO make this listen to Height correctly
	for _, key := range keys {
		value := v.variables[key]
		fmt.Fprintf(&out, "> %s: %s\n", key, value)
	}

	return out.String()
}
