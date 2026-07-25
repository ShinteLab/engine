package main

import (
	"log/slog"
	"os"

	"shinte/engine"
	samples "shinte/engine/_samples"
)

func main() {
	var e samples.ThinkEngine
	err := shogi.StartWithLevel(&e, slog.LevelInfo)
	if err != nil {
		os.Exit(1)
	}
}
