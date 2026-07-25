package main

import (
	"log/slog"
	"os"

	"github.com/ShinteLab/engine"
	samples "github.com/ShinteLab/engine/_samples"
)

func main() {
	var e samples.ThinkEngine
	err := shogi.StartWithLevel(&e, slog.LevelInfo)
	if err != nil {
		os.Exit(1)
	}
}
