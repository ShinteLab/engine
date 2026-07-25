package search

import (
	"context"
	"shinte/engine"
	"testing"
)

// killer/history を使わない(ttBest + 捕獲価値のみの)参照用negamax。
// killer/historyによるノード削減効果を切り分けて確認するためだけの
// テスト専用実装(quiesceは共通で使う)。
func negamaxNoKillerHistory(ctx context.Context, b *shogi.Board, depth, ply, alpha, beta int, nodes *int64, tt *transTable, s *searcher) int {

	*nodes++

	switch b.Repetition() {
	case shogi.RepetitionDraw:
		return 0
	case shogi.RepetitionPerpetualLose:
		return -(MateScore - ply)
	}

	origAlpha := alpha
	var ttBest *shogi.Action
	if tt != nil {
		if score, best, ok := tt.probe(b.Hash(), depth, ply, alpha, beta); ok {
			return score
		} else {
			ttBest = best
		}
	}

	if depth == 0 {
		return s.quiesce(ctx, b, alpha, beta, ply, 0, nodes)
	}

	actions := b.Candidate()
	if len(actions) == 0 {
		return -(MateScore - ply)
	}

	//killer/historyを使わず、ttBest + 捕獲価値降順のみで並べ替える
	orderMoves(actions, ttBest)

	best := -infScore
	var bestMove *shogi.Action
	for _, a := range actions {
		u, ok := b.DoMove(a)
		if !ok {
			continue
		}

		score := -negamaxNoKillerHistory(ctx, b, depth-1, ply+1, -beta, -alpha, nodes, tt, s)

		b.UndoMove(u)

		if score > best {
			best = score
			bestMove = a
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			break
		}
	}

	if tt != nil {
		flag := ttExact
		if best <= origAlpha {
			flag = ttUpper
		} else if best >= beta {
			flag = ttLower
		}
		tt.store(b.Hash(), depth, ply, best, flag, bestMove)
	}

	return best
}

// Stage J-2: killer/history による手順序改善が、(ttBest+捕獲のみの)
// 参照実装と比べて depth4 のノード数を削減することを確認する
// (緩い閾値: 削減されていればよい)。
// 注意: Stage I 時点の約2725ノード(TTのみ、静止探索無し)との単純比較は、
// 静止探索の追加でノード数の絶対値そのものが増えるため意味を持たない。
// ここでは「同じ静止探索を使った上でkiller/historyの有無を切り分ける」
// 比較で効果を検証する。
func TestKillerHistoryReducesNodesAtDepth4(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	ctx := context.Background()
	depth := 4

	full := newSearcher(newTransTable())
	var fullNodes int64
	full.negamax(ctx, b, depth, 0, -infScore, infScore, &fullNodes)

	noKH := newSearcher(newTransTable())
	var noKHNodes int64
	negamaxNoKillerHistory(ctx, b, depth, 0, -infScore, infScore, &noKHNodes, noKH.tt, noKH)

	t.Logf("depth4: full heuristics(killer+history) nodes=%d, without killer/history nodes=%d", fullNodes, noKHNodes)

	if fullNodes >= noKHNodes {
		t.Errorf("expected killer/history move ordering to reduce nodes: with=%d, without=%d", fullNodes, noKHNodes)
	}
}
