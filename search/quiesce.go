package search

import (
	"context"
	"sort"
	"sync/atomic"

	"github.com/ShinteLab/engine"
)

// quiesce内の再帰の最大延長手数(暴走防止)。打ち切り時はevalを返す。
const quiesceMaxPly = 8

// 静止探索。negamax が depth==0 に達した局面で、捕獲の応酬が収まるまで
// (最大 quiesceMaxPly 手)追加で読むことで水平線効果
// (一見駒得だが直後に取り返されて損する手を良い手と誤評価すること)を防ぐ。
//
//   - stand-pat: eval(b) を下限として扱う(手番側は「何もしない」選択肢が
//     常にあるとみなす)。stand-pat が beta 以上なら即カット。
//   - 王手されている場合は stand-pat せず、全合法手で1手延長する
//     (王手放置を防ぐ。合法手が無ければ詰み)。
//   - 王手されていなければ捕獲手のみを価値降順に試す。
//   - qply が quiesceMaxPly に達したら eval を返して打ち切る。
//
// ply は探索ルートからの絶対手数(詰みスコアの ply 補正に使う。
// negamax の ply 引数をそのまま引き継ぐ)。qply は quiesce 内だけの
// 延長手数カウンタ。
func (s *searcher) quiesce(ctx context.Context, b *shogi.Board, alpha, beta, ply, qply int, nodes *int64) int {

	n := atomic.AddInt64(nodes, 1)
	if n%cancelCheckInterval == 0 && ctx.Err() != nil {
		return 0
	}

	inCheck := b.InCheck()

	if !inCheck {
		standPat := eval(b)
		if standPat >= beta {
			return beta
		}
		if standPat > alpha {
			alpha = standPat
		}

		if qply >= quiesceMaxPly {
			return alpha
		}

		actions := b.Candidate()

		captures := make([]*shogi.Action, 0, len(actions))
		for _, a := range actions {
			if a.Enemy() != nil {
				captures = append(captures, a)
			}
		}
		sort.SliceStable(captures, func(i, j int) bool {
			return captureValue(captures[i]) > captureValue(captures[j])
		})

		for _, a := range captures {
			u, ok := b.DoMove(a)
			if !ok {
				continue
			}

			score := -s.quiesce(ctx, b, -beta, -alpha, ply+1, qply+1, nodes)

			b.UndoMove(u)

			if score >= beta {
				return beta
			}
			if score > alpha {
				alpha = score
			}
		}

		return alpha
	}

	//王手中はstand-patせず、全合法手で1手延長する(暴走防止のqply上限も適用)。
	actions := b.Candidate()
	if len(actions) == 0 {
		return -(MateScore - ply)
	}

	if qply >= quiesceMaxPly {
		return eval(b)
	}

	best := -infScore
	for _, a := range actions {
		u, ok := b.DoMove(a)
		if !ok {
			continue
		}

		score := -s.quiesce(ctx, b, -beta, -alpha, ply+1, qply+1, nodes)

		b.UndoMove(u)

		if score > best {
			best = score
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			break
		}
	}

	return best
}
