package search_test

import (
	"shinte/engine"
	"shinte/engine/search"
	"testing"
)

// 水平線効果の実例: 黒飛車が敵歩を取れるが、その歩を守る銀に取り返されて
// 大損する局面(歩+100を得るが飛車-1000を失う)。
// depth=1(この深さでは相手の応手を直接は見ないため、静止探索が無ければ
// 「歩を取って得」という浅い評価のまま採用してしまう)で、
// 静止探索の無い素朴な参照実装(naiveRootBest, search_test.go内)は
// この悪手を選び、静止探索ありの search.Best は正しく回避することを確認する。
func TestQuiesceAvoidsHorizonEffectBlunder(t *testing.T) {

	b, err := shogi.NewBoard("sfen 8K/3s5/4p4/9/4R4/9/9/9/k8 b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	//5e5c(歩を取る)・5e5c+(取って成る)はどちらも、直後に銀に飛車(竜)を
	//取り返されて大損する悪手。(5,3)は成れる位置なので、素朴な参照実装は
	//成りまで込みでさらに得に見える 5e5c+ を選ぶ。
	isBlunder := func(s string) bool {
		return s == "5e5c" || s == "5e5c+"
	}

	//静止探索の無い素朴な参照実装(1手先までしか見ない)は、
	//歩を取れる(+成れる)分だけ得に見えるこの手を選んでしまうはず。
	naiveScore, _ := naiveRootBest(b, 1)
	naiveBestMove := naiveBestAction(b, 1)
	if naiveBestMove == nil || !isBlunder(naiveBestMove.String()) {
		t.Fatalf("setup error: expected naive (no quiescence) reference to pick a blunder (5e5c/5e5c+), got %v (score=%d)",
			naiveBestMove, naiveScore)
	}

	//静止探索ありの本実装は、飛車(竜)を取り返される展開を読んで回避するはず。
	res, err := search.Best(b, search.Options{Depth: 1})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}
	if res.Action == nil {
		t.Fatalf("expected an action, got nil")
	}
	if isBlunder(res.Action.String()) {
		t.Errorf("expected quiescence search to avoid the blunder, but %s was chosen (score=%d)", res.Action.String(), res.Score)
	}
}

// naiveRootBest と同じロジックで、実際に選ばれた手(Action)を返す
// (naiveRootBest はスコアとノード数しか返さないため)。
func naiveBestAction(b *shogi.Board, depth int) *shogi.Action {
	actions := b.Candidate()

	var best *shogi.Action
	bestScore := -(search.MateScore + 1)

	for _, a := range actions {
		nb := b.Copy()
		if !nb.Action(a) {
			continue
		}
		var nodes int64
		score := -naiveNegamax(nb, depth-1, 1, &nodes)
		if best == nil || score > bestScore {
			bestScore = score
			best = a
		}
	}
	return best
}
