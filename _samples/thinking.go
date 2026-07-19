package samples

import (
	"context"
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

// StageD: 置換表(TT)を有効化。depth4 でも単一スレッド+TTで
// 数十ms程度で返る(ベンチ計測済み)ため depth を 3->4 に引き上げた。
// StageK: 置換表がロックフリー(xor-verify)方式で共有可能になり、
// Lazy SMP(Parallel=true)が全workerで共有TT1枚のみを使うようになった
// ため、StageD時点の「Parallel+TTでGB級メモリ」の問題は解消された。
// Workers:0(自動: runtime.NumCPU()と8の小さい方)で並列探索を有効化する。
const thinkDepth = 4

func (e *ThinkEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {
	return e.GetBestContext(context.Background(), b, nil)
}

// StageF: ContextEngine 対応。ctx でのキャンセル・info コールバックの
// 転送に対応する。
func (e *ThinkEngine) GetBestContext(ctx context.Context, b *shogi.Board, info func(shogi.Info)) (*shogi.Action, error) {
	e.depth = thinkDepth
	slog.Info(fmt.Sprintf("Think[%d]", e.depth))

	opt := search.Options{Depth: e.depth, Parallel: true, Workers: 0, TT: true}
	if info != nil {
		opt.Info = func(si search.Info) {
			info(shogi.Info{
				Depth:   si.Depth,
				ScoreCP: si.ScoreCP,
				Nodes:   si.Nodes,
				PV:      si.PV,
			})
		}
	}

	res, err := search.BestContext(ctx, b, opt)
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

// StageG: 詰将棋探索(go mate)対応。
const mateMaxDepth = 9

func (e *ThinkEngine) GetMate(ctx context.Context, b *shogi.Board) ([]*shogi.Action, error) {
	moves, found, err := search.Mate(ctx, b, mateMaxDepth)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return moves, nil
}
