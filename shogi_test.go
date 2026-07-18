package shogi_test

import (
	"log/slog"
	"os"
	"shogi"
	"testing"
)

func TestMain(m *testing.M) {
	defer shogi.SetFileLogger(slog.LevelDebug, "_dist/board_test.log", false).Close()
	code := m.Run()
	os.Exit(code)
}

func parse(line string) *shogi.Board {
	b, err := shogi.NewBoard(line)
	if err != nil {
		panic(err)
	}
	return b
}
