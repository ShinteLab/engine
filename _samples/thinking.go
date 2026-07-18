package samples

import (
	"fmt"
	"log/slog"

	"shogi"
	"shogi/search"
)

type ThinkEngine struct {
	depth int
}

func (e *ThinkEngine) GetName() string {
	return "thinking Engine"
}

func (e *ThinkEngine) GetVersion() string {
	return "0.0.0"
}

func (e *ThinkEngine) GetAuthor() string {
	return "secondarykey"
}

func (e *ThinkEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {
	e.depth = 4
	slog.Info(fmt.Sprintf("Think[%d]", e.depth))

	// StageD: 置換表(TT)を有効化。depth4 でも単一スレッド+TTで
	// 数十ms程度で返る(ベンチ計測済み)ため depth を 3->4 に引き上げた。
	// Parallel(ルート並列)は TT と併用すると反復深化の各深さ×ルート手数分の
	// TT インスタンスが goroutine ごとに複製され、depth4 で数GB級のメモリ
	// 確保が発生し単一スレッドより遅くなることを計測で確認したため、
	// 共有TTが無い現状の設計では Parallel=false + TT=true を採用する。
	res, err := search.Best(b, search.Options{Depth: e.depth, Parallel: false, TT: true})
	if err != nil {
		if err == search.ErrNoMoves {
			//合法手が無い(詰み)ので投了
			slog.Info("Think End (resign)")
			return shogi.ResignAction(), nil
		}
		return nil, err
	}

	slog.Info(fmt.Sprintf("Think End Score:%d Nodes:%d", res.Score, res.Nodes))
	return res.Action, nil
}
