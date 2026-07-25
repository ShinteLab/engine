package main

import (
	"os"

	"github.com/ShinteLab/engine"
	samples "github.com/ShinteLab/engine/_samples"
)

func main() {
	var e samples.GiveupEngine
	err := shogi.Start(&e)
	if err != nil {
		os.Exit(1)
	}
}
