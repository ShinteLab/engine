package shogi

import (
	"log/slog"
	"os"
)

func Start(engine Engine) error {
	return StartWithLevel(engine, slog.LevelInfo)
}

func StartWithLevel(engine Engine, lv slog.Level) error {
	return StartWithLogFile(engine, "shogi_%d.log", lv)
}

func StartWithLogFile(engine Engine, name string, lv slog.Level) error {
	defer SetFileLogger(lv, name, false).Close()
	usi := NewUSI(os.Stdout, os.Stdin, engine)
	return usi.Start()
}
