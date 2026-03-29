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
	Cursor        int
	offset        int

	breakpoints map[int]struct{}

	currStoppedLine int
}

func NewCodeViewer(text string) CodeViewer {
	cv := CodeViewer{
		Width:           0,
		Height:          0,
		Cursor:          0,
		offset:          0,
		currStoppedLine: -1,
		breakpoints:     map[int]struct{}{},
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
	case ChangeTextMsg:
		c = c.colorize(msg.Text)
	case BreakpointMsg:
		if msg.Added {
			c.breakpoints[msg.Line] = struct{}{}
		} else {
			delete(c.breakpoints, msg.Line)
		}
	case ExecutionStoppedMsg:
		c.currStoppedLine = msg.Line
	}

	return c, nil
}

func (c CodeViewer) handleKeyMsg(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "k":
		c.Cursor = max(0, c.Cursor-1)
		topThreshold := c.offset + int(float64(c.Height)*0.1)
		if c.Cursor < topThreshold && c.offset > 0 {
			c.offset -= 1
		}
	case "j":
		c.Cursor = min(len(c.lines)-1, c.Cursor+1)
		bottomThreshold := c.offset + int(float64(c.Height)*0.9)
		if c.Cursor > bottomThreshold && c.offset < len(c.lines) {
			c.offset += 1
		}
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
		lineNumber := i + c.offset + 1
		breakpoint := " "
		cursor := " "
		executor := " "

		if lineNumber == c.Cursor+1 {
			cursor = ">"
		}
		if _, ok := c.breakpoints[lineNumber]; ok {
			breakpoint = "O"
		}
		if lineNumber == c.currStoppedLine {
			executor = "@"
		}

		fmt.Fprintf(&out, "%-4d%s%s%s %s\n", lineNumber, breakpoint, executor, cursor, line)
	}
	// Count long lines, and add here if they are out of view, this is a bit tricky I guess
	if len(lines)-c.offset < c.Height {
		for range c.Height - len(lines) {
			fmt.Fprintf(&out, "\n")
		}
	}

	return out.String()
}
