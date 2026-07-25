package main

import (
	"os"

	"shinte/engine"
	samples "shinte/engine/_samples"
)

func main() {
	var e samples.GiveupEngine
	err := shogi.Start(&e)
	if err != nil {
		os.Exit(1)
	}
}
