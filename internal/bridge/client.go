package bridge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"slices"
	"sync"
)

var ErrNotRunning = errors.New("bridge not running")
var ErrAlreadyStarted = errors.New("bridge is already running")

type ExecutionPoint struct {
	File string
	Line int
}

type Bridge struct {
	stdin  io.Writer
	stdout *bufio.Reader

	cmd *exec.Cmd

	input chan string

	outputLock *sync.RWMutex
	output     []chan string

	executionStopLock *sync.RWMutex
	executionStop     []chan ExecutionPoint

	registry     map[string]chan string
	registryLock *sync.RWMutex

	breakpoints map[string][]int

	callbacksLock *sync.RWMutex
	callbacks     map[string]map[int]func()

	running bool

	path string
}

func NewBridge(path string) *Bridge {
	return &Bridge{
		path:         path,
		input:        make(chan string),
		registry:     make(map[string]chan string),
		registryLock: &sync.RWMutex{},
		breakpoints:  make(map[string][]int),

		callbacksLock: &sync.RWMutex{},
		callbacks:     make(map[string]map[int]func()),

		output:     make([]chan string, 0),
		outputLock: &sync.RWMutex{},

		executionStop:     make([]chan ExecutionPoint, 0),
		executionStopLock: &sync.RWMutex{},

		running: false,
	}
}

func (b *Bridge) Start() error {
	if b.running {
		return ErrAlreadyStarted
	}

	b.cmd = exec.Command(
		"python",
		"-u",
		"python/pybug_runtime.py",
		b.path,
	)

	var err error
	b.stdin, err = b.cmd.StdinPipe()
	if err != nil {
		return err
	}

	reader, err := b.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	b.stdout = bufio.NewReader(reader)

	err = b.cmd.Start()
	if err != nil {
		return err
	}

	b.running = true
	go b.readLoop()
	go b.writeLoop()

	return nil
}

func (b *Bridge) Subscribe() chan string {
	b.outputLock.Lock()
	defer b.outputLock.Unlock()

	c := make(chan string)
	b.output = append(b.output, c)

	return c
}

func (b *Bridge) SubscribeStopped() chan ExecutionPoint {
	b.executionStopLock.Lock()
	defer b.executionStopLock.Unlock()

	c := make(chan ExecutionPoint)
	b.executionStop = append(b.executionStop, c)

	return c
}

func (b *Bridge) Locals() (map[string]any, error) {
	if !b.running {
		return nil, ErrNotRunning
	}

	requestId, cmd := makeCommand(LocalsCommand, map[string]any{})

	c := b.sendCommand(requestId, cmd)

	obj := <-c

	var m map[string]any
	err := json.Unmarshal([]byte(obj), &m)
	if err != nil {
		return nil, err
	}

	vars, ok := m["vars"].(map[string]any)
	if !ok {
		return nil, errors.New("could not extract vars from response")
	}

	return vars, nil
}

func (b *Bridge) Step() error {
	if !b.running {
		return ErrNotRunning
	}

	_, cmd := makeCommand(StepCommand, nil)
	b.sendCommandNoResponse(cmd)

	return nil
}

func (b *Bridge) Breakpoint(file string, line int) (set bool, err error) {
	if !b.running {
		return false, ErrNotRunning
	}

	var command CommandType
	if _, ok := b.breakpoints[file]; ok && slices.Contains(b.breakpoints[file], line) {
		command = UnbreakCommand
	} else {
		command = BreakCommand
	}

	requestId, cmd := makeCommand(command, map[string]any{
		"file": file,
		"line": line,
	})

	c := b.sendCommand(requestId, cmd)

	obj := <-c

	var m map[string]any
	err = json.Unmarshal([]byte(obj), &m)
	if err != nil {
		return false, err
	}

	if m["status"] != "ok" {
		return false, fmt.Errorf("error occured on break, err: %s", m["error"])
	}

	if command == BreakCommand {
		b.breakpoints[file] = append(b.breakpoints[file], line)
		return true, nil
	} else {
		breakpointsLen := len(b.breakpoints[file])
		index := slices.Index(b.breakpoints[file], line)
		b.breakpoints[file][index] = b.breakpoints[file][breakpointsLen-1]

		b.breakpoints[file] = b.breakpoints[file][0 : breakpointsLen-1]
		return false, nil
	}
}

func (b *Bridge) Continue() error {
	if !b.running {
		return ErrNotRunning
	}

	_, cmd := makeCommand(ContinueCommand, map[string]any{})

	b.sendCommandNoResponse(cmd)
	return nil
}

func (b *Bridge) sendCommandNoResponse(command string) {
	b.input <- command
}

func (b *Bridge) sendCommand(requestId string, command string) chan string {
	b.registryLock.Lock()
	defer b.registryLock.Unlock()

	channel := make(chan string)

	b.registry[requestId] = channel
	b.input <- command

	return channel
}

func (b *Bridge) writeLoop() {
	slog.Debug("started writeLoop")
	defer slog.Debug("writeLoop exited")
	for {
		cmd := <-b.input

		if !b.running {
			return
		}

		slog.Info("Received command", "cmd", cmd)

		_, err := b.stdin.Write([]byte(cmd + "\n"))
		if err != nil {
			slog.Error("Error occured while writing to stdin", "error", err)
		}

		slog.Debug("Command written", "cmd", cmd)
	}
}

func (b *Bridge) readLoop() {
	slog.Debug("started readLoop")
	defer slog.Debug("readLoop exited")
	for {
		slog.Debug("reading string from stdout waiting for newline")
		line, err := b.stdout.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			slog.Error("Error occured while reading from stdout", "error", err)
			continue
		} else if errors.Is(err, io.EOF) {
			b.running = false
			return
		}

		var msg map[string]any
		err = json.Unmarshal([]byte(line), &msg)
		if err != nil {
			slog.Debug("read line from stdout", "line", line)

			b.outputLock.RLock()
			for _, c := range b.output {
				c <- line
			}
			b.outputLock.RUnlock()

		} else if requestId, ok := msg["request_id"].(string); ok {
			b.registryLock.RLock()
			c, ok := b.registry[requestId]
			b.registryLock.RUnlock()

			if !ok {
				slog.Error("Could not find requestId in registry", "requestId", requestId)
				continue
			}

			c <- line
		} else {
			b.handleStopped(msg)
		}
	}
}

func (b *Bridge) handleStopped(msg map[string]any) error {
	// TODO: set to stopped
	if event, ok := msg["event"]; !ok || event != "stopped" {
		slog.Warn("received unkown event", "msg", msg)
		return errors.New("unknown event encountered")
	}

	file := msg["file"].(string)
	line, ok := toInt(msg["line"])
	if !ok {
		slog.Error("could not convert line to int", "line", msg["line"])
		return errors.New("could not convert line to int")
	}

	slog.Info("received stopped event")
	b.callbacksLock.RLock()
	defer b.callbacksLock.RUnlock()

	if callback, ok := b.callbacks[file][line]; ok {
		slog.Info("found callback, now running", "file", file, "line", line)
		go callback()
	}

	b.executionStopLock.RLock()
	defer b.executionStopLock.RUnlock()

	for _, c := range b.executionStop {
		c <- ExecutionPoint{
			File: file,
			Line: line,
		}
	}

	return nil
}

func (b *Bridge) OnBreakpoint(file string, line int, callback func()) {
	b.callbacksLock.Lock()
	defer b.callbacksLock.Unlock()

	if f, ok := b.callbacks[file]; ok {
		f[line] = callback
	} else {
		b.callbacks[file] = map[int]func(){
			line: callback,
		}
	}
}

func (b *Bridge) Wait() error {
	return b.cmd.Wait()
}

func toInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int8:
		return int(x), true
	case int16:
		return int(x), true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float32:
		return int(x), true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}
