package codeviewer

type ChangeTextMsg struct {
	Text string
}

type BreakpointMsg struct {
	Line  int
	Added bool
}

type ExecutionStoppedMsg struct {
	Line int
}
