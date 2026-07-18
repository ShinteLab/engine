package search_test

import (
	"shogi"
	"shogi/search"
	"testing"
)

// 1手詰み: 白玉を金打ちで詰ます局面。Best() が詰み手を返し、
// Score が MateScore 級であることを確認する。
func TestBestFindsMateInOne(t *testing.T) {

	b, err := shogi.NewBoard("sfen 3pkp3/3s1s3/4S4/9/9/9/9/9/9 b G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	res, err := search.Best(b, search.Options{Depth: 3})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}

	if res.Action == nil {
		t.Fatalf("expected an action, got nil")
	}
	if res.Action.String() != "G*5b" {
		t.Errorf("want mate move G*5b, got %s (score=%d nodes=%d)", res.Action.String(), res.Score, res.Nodes)
	}
	if res.Score < search.MateScore-100 {
		t.Errorf("expected mate-level score, got %d", res.Score)
	}
}

// 駒得選択: ただで敵飛車が取れる局面で、捕獲手を選ぶことを確認する。
func TestBestPrefersFreeCapture(t *testing.T) {

	b, err := shogi.NewBoard("sfen k8/9/4r4/9/4R4/9/9/9/8K b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	res, err := search.Best(b, search.Options{Depth: 2})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}

	// (5,3)は成れる位置なので、飛車を成って取る5e5c+の方が価値が高く選ばれる。
	// どちらも「捕獲手」であることを検証すればよい。
	if res.Action == nil {
		t.Fatalf("expected an action, got nil")
	}
	s := res.Action.String()
	if s != "5e5c" && s != "5e5c+" {
		t.Errorf("want capturing move 5e5c or 5e5c+, got %v (score=%d)", res.Action, res.Score)
	}
	if res.Action.Enemy() == nil {
		t.Errorf("expected the chosen move to be a capture: %v", res.Action)
	}
}

// 王手回避: 王手されている局面で、返された手が実際に王手を解消することを
// E2E で確認する(Candidate() が合法手のみを返すため自然に満たされるはずだが
// 探索経路として確認する)。
func TestBestEscapesCheck(t *testing.T) {

	b, err := shogi.NewBoard("sfen 4r4/9/9/9/4K4/9/9/9/8k b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if !b.IsCheck(shogi.TurnBlack) {
		t.Fatalf("expected black to be in check")
	}

	res, err := search.Best(b, search.Options{Depth: 2})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}
	if res.Action == nil {
		t.Fatalf("expected an evasion action, got nil")
	}

	nb := b.Copy()
	if !nb.Action(res.Action) {
		t.Fatalf("action %s failed to apply", res.Action.String())
	}
	if nb.IsCheck(shogi.TurnBlack) {
		t.Errorf("returned action %s does not escape check", res.Action.String())
	}
}

// ---- αβの健全性確認用: 枝刈りを行わない素朴な参照実装 ----

func naiveEval(b *shogi.Board) int {
	t := b.Turn()
	return b.Material(t) - b.Material(t.Next())
}

func naiveNegamax(b *shogi.Board, depth, ply int, nodes *int64) int {
	*nodes++

	if depth == 0 {
		return naiveEval(b)
	}

	actions := b.Candidate()
	if len(actions) == 0 {
		return -(search.MateScore - ply)
	}

	best := -(search.MateScore + 1)
	for _, a := range actions {
		nb := b.Copy()
		if !nb.Action(a) {
			continue
		}
		score := -naiveNegamax(nb, depth-1, ply+1, nodes)
		if score > best {
			best = score
		}
	}
	return best
}

// ルートの全手を(枝刈りなしで)展開し、最善値とノード数を返す。
func naiveRootBest(b *shogi.Board, depth int) (int, int64) {
	actions := b.Candidate()

	var nodes int64
	best := -(search.MateScore + 1)
	for _, a := range actions {
		nb := b.Copy()
		if !nb.Action(a) {
			continue
		}
		score := -naiveNegamax(nb, depth-1, 1, &nodes)
		if score > best {
			best = score
		}
	}
	return best, nodes
}

// αβの健全性: 同一局面・同一深さで、αβありの探索と全展開の参照実装の
// ルート評価値が一致することを確認する。
func TestAlphaBetaMatchesNaiveRootValue(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	depth := 2

	abRes, err := search.Best(b, search.Options{Depth: depth})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}

	naiveScore, _ := naiveRootBest(b, depth)

	if abRes.Score != naiveScore {
		t.Errorf("alpha-beta root score %d != naive root score %d", abRes.Score, naiveScore)
	}
}

// ノード数削減: depth3 で αβ の Nodes が参照実装の Nodes より大幅に
// (おおむね半分以下)小さいことを確認する。
func TestAlphaBetaReducesNodes(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	depth := 3

	abRes, err := search.Best(b, search.Options{Depth: depth, Parallel: false})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}

	_, naiveNodes := naiveRootBest(b, depth)

	t.Logf("alpha-beta nodes=%d naive nodes=%d", abRes.Nodes, naiveNodes)

	if abRes.Nodes >= naiveNodes/2 {
		t.Errorf("expected alpha-beta nodes (%d) to be well under half of naive nodes (%d)", abRes.Nodes, naiveNodes)
	}
}

// BenchmarkSearchDepth3: 初期局面での depth3 探索(単一スレッド)の速度計測。
func BenchmarkSearchDepth3(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		search.Best(board, search.Options{Depth: 3, Parallel: false})
	}
}

// 置換表 on/off で、同一局面・同一深さのルート最善手のスコアが一致することを
// 確認する(手そのものは同点別手を許容し、スコアで比較する)。
func TestTTOnOffScoreMatches(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	depth := 3

	withTT, err := search.Best(b, search.Options{Depth: depth, TT: true})
	if err != nil {
		t.Fatalf("Best() with TT error: %v", err)
	}

	withoutTT, err := search.Best(b, search.Options{Depth: depth, TT: false})
	if err != nil {
		t.Fatalf("Best() without TT error: %v", err)
	}

	if withTT.Score != withoutTT.Score {
		t.Errorf("TT on score %d != TT off score %d", withTT.Score, withoutTT.Score)
	}
}

// 置換表 on で depth4 の Nodes が TT off より減ることを確認する。
func TestTTReducesNodesAtDepth4(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	depth := 4

	withTT, err := search.Best(b, search.Options{Depth: depth, TT: true})
	if err != nil {
		t.Fatalf("Best() with TT error: %v", err)
	}

	withoutTT, err := search.Best(b, search.Options{Depth: depth, TT: false})
	if err != nil {
		t.Fatalf("Best() without TT error: %v", err)
	}

	t.Logf("depth4 TT on nodes=%d, TT off nodes=%d", withTT.Nodes, withoutTT.Nodes)

	if withTT.Nodes >= withoutTT.Nodes {
		t.Errorf("expected TT-on nodes (%d) to be less than TT-off nodes (%d)", withTT.Nodes, withoutTT.Nodes)
	}
}

// go test -race ./search が通ること(Parallel=true + TT=true の組み合わせも
// 含めてレースが無いことを確認する)。
func TestParallelWithTTNoRace(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	res, err := search.Best(b, search.Options{Depth: 3, Parallel: true, TT: true})
	if err != nil {
		t.Fatalf("Best() error: %v", err)
	}
	if res.Action == nil {
		t.Fatalf("expected an action, got nil")
	}
}

// BenchmarkSearchDepth4TT: depth4・置換表ありの速度計測。
func BenchmarkSearchDepth4TT(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		search.Best(board, search.Options{Depth: 4, Parallel: false, TT: true})
	}
}

// BenchmarkSearchDepth4NoTT: depth4・置換表なしの速度計測。
func BenchmarkSearchDepth4NoTT(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		search.Best(board, search.Options{Depth: 4, Parallel: false, TT: false})
	}
}
