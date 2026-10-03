package shogi

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

// ログ。engine はライブラリなので、ログの設定（レベル・出力先）は持たない。
// 使う側が *slog.Logger を SetLogger で渡す。渡さなければ slog.Default() に出す。
//
// ⚠️ 標準出力へ書く Logger を渡さないこと。Start は USI のやり取りに標準入出力を使うので、
// ログが混ざると GUI が USI の応答として読んでしまう（slog.Default() の既定の出口は標準エラー）。

var pkgLogger atomic.Pointer[slog.Logger]

// SetLogger は engine が使う Logger を差し替える。nil を渡すと slog.Default() に戻る。
// レベルを変えたいときは、絞ったハンドラで作った Logger を渡す。
func SetLogger(l *slog.Logger) {
	pkgLogger.Store(l)
}

// logger は今の Logger を返す。slog.Default() は呼ぶたびに引く（使う側があとから
// slog.SetDefault したときに追従するため）。
func logger() *slog.Logger {
	if l := pkgLogger.Load(); l != nil {
		return l
	}
	return slog.Default()
}

// openLogFile は Start が使うログファイルを作る。name の "%d" はプロセス ID に置き換える。
func openLogFile(name string, lv slog.Level) (*slog.Logger, io.Closer, error) {
	if strings.Contains(name, "%d") {
		name = fmt.Sprintf(name, os.Getpid())
	}
	fp, err := os.Create(name)
	if err != nil {
		return nil, nil, err
	}
	h := slog.NewTextHandler(fp, &slog.HandlerOptions{Level: lv})
	return slog.New(h), fp, nil
}
