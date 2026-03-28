package main

import (
	"log/slog"
	"os"

	b "git.pablu.de/pablu/pybug/internal/bridge"
	"git.pablu.de/pablu/pybug/ui"
)

func main() {
	f, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	handler := slog.NewTextHandler(f, &slog.HandlerOptions{
		AddSource: true,
	})

	slog.SetDefault(slog.New(handler))
	slog.SetLogLoggerLevel(slog.LevelDebug)
	bridge := b.NewBridge("test.py")

	err = ui.Run(bridge)
	if err != nil {
		panic(err)
	}
}
