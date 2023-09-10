package shogi

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func SetLog(h slog.Handler) error {
	slog.SetDefault(slog.New(h))
	return nil
}

type closer struct {
	w io.Closer
}

func SetFileLogger(lv slog.Level, name string) *closer {

	wk := name
	if strings.Index(wk, "%d") != -1 {
		pid := os.Getpid()
		wk = fmt.Sprintf(wk, pid)
	}

	fp, err := os.Create(wk)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	var opts slog.HandlerOptions
	opts.Level = lv
	h := slog.NewTextHandler(fp, &opts)
	SetLog(h)

	var rtn closer
	rtn.w = fp
	return &rtn
}

func (c *closer) Close() {
	if c != nil {
		c.w.Close()
	}
}
