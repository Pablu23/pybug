package main

import (
	// "fmt"
	// "log/slog"
	// "time"

	"log/slog"

	b "git.pablu.de/pablu/pybug/internal/bridge"
	"git.pablu.de/pablu/pybug/ui"
)

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	//
	// fmt.Println("Started bridge")
	//
	// err = bridge.Breakpoint("test.py", 5)
	// bridge.OnBreakpoint("test.py", 5, func() {
	// 	locals, err := bridge.Locals()
	// 	if err != nil {
	// 		slog.Error("Encountered error on callback", "error", err)
	// 		return
	// 	}
	//
	// 	for key, val := range locals {
	// 		slog.Info("found local variable", "key", key, "value", val)
	// 	}
	// })
	//
	// bridge.Continue()
	//
	// time.Sleep(5 * time.Second)
	//
	// bridge.Continue()
	//
	// err = bridge.Wait()
	// if err != nil {
	// 	panic(err)
	// }

	slog.SetLogLoggerLevel(slog.LevelError)

	bridge := b.NewBridge("test.py")
	err := bridge.Start()
	if err != nil {
		panic(err)
	}

	err = ui.Run(bridge)
	if err != nil {
		panic(err)
	}
}
