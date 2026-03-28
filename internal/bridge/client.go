package bridge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
)

type Bridge struct {
	stdin  io.Writer
	stdout *bufio.Reader

	cmd *exec.Cmd

	input chan string

	outputLock *sync.RWMutex
	output     []chan string

	registry     map[string]chan string
	registryLock *sync.RWMutex

	breakpoints map[string][]int

	callbacksLock *sync.RWMutex
	callbacks     map[string]map[int]func()
}

func NewBridge(path string) *Bridge {
	cmd := exec.Command(
		"python",
		"-u",
		"python/pybug_runtime.py",
		path,
	)

	return &Bridge{
		cmd:          cmd,
		input:        make(chan string),
		registry:     make(map[string]chan string),
		registryLock: &sync.RWMutex{},
		breakpoints:  make(map[string][]int),

		callbacksLock: &sync.RWMutex{},
		callbacks:     make(map[string]map[int]func()),

		output:     make([]chan string, 0),
		outputLock: &sync.RWMutex{},
	}
}

func (b *Bridge) Start() error {
	var err error
	b.stdin, err = b.cmd.StdinPipe()
	if err != nil {
		return err
	}
	// b.cmd.Stdout = os.Stdout

	reader, err := b.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	b.stdout = bufio.NewReader(reader)

	err = b.cmd.Start()
	if err != nil {
		return err
	}

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

func (b *Bridge) Locals() (map[string]any, error) {
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

func (b *Bridge) Breakpoint(file string, line int) error {
	// Check if breakpoint already exists here

	requestId, cmd := makeCommand(BreakCommand, map[string]any{
		"file": file,
		"line": line,
	})

	c := b.sendCommand(requestId, cmd)

	obj := <-c

	var m map[string]any
	err := json.Unmarshal([]byte(obj), &m)
	if err != nil {
		return err
	}

	if m["status"] != "ok" {
		return fmt.Errorf("error occured on break, err: %s", m["error"])
	}

	b.breakpoints[file] = append(b.breakpoints[file], line)

	return nil
}

func (b *Bridge) Continue() {
	_, cmd := makeCommand(ContinueCommand, map[string]any{})

	b.sendCommandNoResponse(cmd)
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
			// TODO: set to stopped
			if event, ok := msg["event"]; !ok || event != "stopped" {
				slog.Warn("received unkown event", "msg", msg)
			}

			file := msg["file"].(string)
			line, ok := toInt(msg["line"])
			if !ok {
				slog.Error("could not convert line to int", "line", msg["line"])
			}

			slog.Info("received stopped event")
			b.callbacksLock.RLock()
			if callback, ok := b.callbacks[file][line]; ok {
				slog.Info("found callback, now running", "file", file, "line", line)
				go callback()
			}
			b.callbacksLock.RUnlock()
		}
	}
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
