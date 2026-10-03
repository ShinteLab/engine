package shogi_test

import (
	"log/slog"
	"os"
	"github.com/ShinteLab/engine"
	"testing"
)

func TestMain(m *testing.M) {
	// engine のログは _dist/board_test.log へ（_dist が無ければ標準エラーのまま）。
	if fp, err := os.Create("_dist/board_test.log"); err == nil {
		shogi.SetLogger(slog.New(slog.NewTextHandler(fp, &slog.HandlerOptions{Level: slog.LevelDebug})))
		code := m.Run()
		fp.Close()
		os.Exit(code)
	}
	os.Exit(m.Run())
}

func parse(line string) *shogi.Board {
	b, err := shogi.NewBoard(line)
	if err != nil {
		panic(err)
	}
	return b
}
