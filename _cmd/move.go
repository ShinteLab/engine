package main

import (
	"log/slog"
	"os"

	"shogi"
	samples "shogi/_samples"
)

func main() {
	var e samples.MoveEngine
	err := shogi.StartWithLevel(&e, slog.LevelDebug)
	if err != nil {
		os.Exit(1)
	}
}
