package main

import (
	"os"

	"shogi"
	"shogi/samples"
)

func main() {
	var e samples.MoveEngine
	err := shogi.Start(&e)
	if err != nil {
		os.Exit(1)
	}
}
