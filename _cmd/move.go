package main

import (
	"log/slog"
	"os"

	"shinte/engine"
	samples "shinte/engine/_samples"
)

func main() {
	var e samples.MoveEngine
	err := shogi.StartWithLevel(&e, slog.LevelDebug)
	if err != nil {
		os.Exit(1)
	}
}
