package codeviewer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type CodeViewer struct {
	lines []string

	Width, Height int
	cursor        int
	offset        int
}

func NewCodeViewer(text string) CodeViewer {
	cv := CodeViewer{
		Width:  0,
		Height: 0,
		cursor: 0,
		offset: 0,
	}

	return cv.colorize(text)
}

func (c CodeViewer) Init() tea.Cmd {
	return nil
}

func (c CodeViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return c.handleKeyMsg(msg)
		// case tea.WindowSizeMsg:
		// return c.UpdateWindowSize(msg)
	}

	return c, nil
}

func (c CodeViewer) handleKeyMsg(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "k":
		c.cursor = max(0, c.cursor-1)
		topThreshold := c.offset + int(float64(c.Height)*0.1)
		if c.cursor < topThreshold && c.offset > 0 {
			c.offset -= 1
		}
	case "j":
		c.cursor = min(len(c.lines)-1, c.cursor+1)
		bottomThreshold := c.offset + int(float64(c.Height)*0.9)
		if c.cursor > bottomThreshold && c.offset < len(c.lines) {
			c.offset += 1
		}

		// case "b":
		// 	lineNumber := c.cursor + 1
		//
		// 	c.bridge.Breakpoint(m.currentFile, lineNumber)
		// 	if file, ok := m.breakpoints[m.currentFile]; ok {
		// 		m.breakpoints[m.currentFile] = append(file, lineNumber)
		// 	} else {
		// 		m.breakpoints[m.currentFile] = []int{
		// 			lineNumber,
		// 		}
		// 	}
		// case "s":
		// 	m.bridge.Step()
		// case "c":
		// 	m.bridge.Continue()
		// case "r":
		// 	m.messages = make([]string, 0)
		// 	m.stdoutOutput.SetContent("")
		// 	err := m.bridge.Start()
		// 	if err != nil {
		// 		slog.Error("could not start brige", "error", err)
		// 	}
	}

	return c, nil
}

func (c CodeViewer) colorize(text string) CodeViewer {
	var buf bytes.Buffer
	lexer := lexers.Get("python")
	style := styles.Get("monokai")
	formatter := formatters.Get("terminal16m")

	iterator, _ := lexer.Tokenise(nil, text)
	formatter.Format(&buf, style, iterator)

	c.lines = strings.Split(buf.String(), "\n")
	return c
}

func (c CodeViewer) View() string {
	var out strings.Builder

	lines := c.lines[max(0, c.offset):min(c.offset+c.Height, len(c.lines))]
	for i, line := range lines {
		breakpoint := " "
		cursor := " "
		executor := " "
		if c.offset+i == c.cursor {
			cursor = ">"
		}
		fmt.Fprintf(&out, "%-4d%s%s%s %s\n", c.offset+i+1, breakpoint, executor, cursor, line)
	}
	if len(lines)-c.offset < c.Height {
		for range c.Height - len(lines) {
			fmt.Fprintf(&out, "\n")
		}
	}

	return out.String()
}
