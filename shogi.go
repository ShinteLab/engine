package shogi

import (
	"log/slog"
	"os"
)

// Start / StartWithLevel / StartWithLogFile は、エンジンの**実行ファイルの main から呼ぶ入口**。
// プロセスを USI エンジンにする（標準入力で USI のコマンドを受け、標準出力へ応答する）。
//
// プロセスをまるごと預かる入口なので、ログもここで決める: ファイルを作り、
// slog の既定の Logger をそこへ向ける（engine 自身のログも、Engine の実装が slog に
// 直接出すログも、同じファイルに入る）。終わったら元に戻す。
//
// ⚠️ ライブラリとして組み込むとき（同じプロセスの中で USI を繋ぐなど）はこれを使わず、
// NewUSI と SetLogger を使うこと。既定の Logger を書き換えてしまう。

func Start(engine Engine) error {
	return StartWithLevel(engine, slog.LevelInfo)
}

func StartWithLevel(engine Engine, lv slog.Level) error {
	return StartWithLogFile(engine, "shogi_%d.log", lv)
}

// StartWithLogFile は name のファイルへ lv 以上のログを書く（"%d" はプロセス ID）。
// ファイルを作れなくてもエンジンは動かす（ログは標準エラーへ出る）。
func StartWithLogFile(engine Engine, name string, lv slog.Level) error {
	if l, c, err := openLogFile(name, lv); err != nil {
		logger().Error("ログファイルを作れません", "name", name, "err", err)
	} else {
		prev := slog.Default()
		slog.SetDefault(l)
		defer func() {
			slog.SetDefault(prev)
			c.Close()
		}()
	}
	usi := NewUSI(os.Stdout, os.Stdin, engine)
	return usi.Start()
}
