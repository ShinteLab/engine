package main

import (
	"log/slog"
	"os"

	"shogi"
	samples "shogi/_samples"
)

func main() {
	var e samples.ThinkEngine
	err := shogi.StartWithLevel(&e, slog.LevelInfo)
	if err != nil {
		os.Exit(1)
	}
}
