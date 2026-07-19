package search_test

import (
	"context"
	"shogi"
	"shogi/search"
	"testing"
	"time"
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

// Movetime を短く指定すると即座に打ち切られ、緩い時間内に応答が返ることを確認する。
func TestBestContextMovetimeCancelsQuickly(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	start := time.Now()
	res, err := search.BestContext(context.Background(), b, search.Options{Depth: 20, Movetime: 10 * time.Millisecond})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("BestContext() error: %v", err)
	}
	if res.Action == nil {
		t.Fatalf("expected a valid action even when cancelled")
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("expected cancellation to return quickly, took %v", elapsed)
	}
}

// Movetime を指定しない場合は従来どおり指定深さまで完走することを確認する。
func TestBestContextNoMovetimeCompletesDepth(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	res, err := search.BestContext(context.Background(), b, search.Options{Depth: 2})
	if err != nil {
		t.Fatalf("BestContext() error: %v", err)
	}
	if res.Action == nil {
		t.Fatalf("expected a valid action")
	}
}

// 既にキャンセル済みの ctx を渡しても、必ず有効な手(先頭手)を返すことを確認する。
func TestBestContextAlreadyCancelledStillReturnsMove(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := search.BestContext(ctx, b, search.Options{Depth: 3})
	if err != nil {
		t.Fatalf("BestContext() error: %v", err)
	}
	if res.Action == nil {
		t.Fatalf("expected a fallback action even for an already-cancelled context")
	}
}

// BenchmarkSearchDepth3: 初期局面での depth3 探索(単一スレッド)の速度計測。
// StageC: 2265433 ns/op(2.27ms)
// StageD: 2815844 ns/op(zobrist hash追加)
// StageE: 3534559 ns/op相当(history追加、未計測のため後日F-0時に実測)
// StageF-0: legalCandidate内をcopyLite/actionLite化してCandidate()呼び出し
// コストは削減したが、negamax自身のCopy+ActionはRepetition()判定のため
// 引き続き履歴付きの正規Copy/Actionを使うので完全には2.3ms水準へは戻らない。
// BenchmarkSearchDepth3-20            416           3534559 ns/op
// StageF(仕上げ再計測): 397           3274457 ns/op
// StageH: legalCandidateをピン検出方式化(shogi側)、negamax内のCopy+Action
// が呼ぶCandidate()が軽量化されたことで2.3ms水準を超えて短縮
// BenchmarkSearchDepth3-20            930           1527033 ns/op
// StageI: negamax/searchRootをCopy+ActionからDoMove/UndoMoveへ変更
// (単一Boardを掘り下げる)。allocs/opもほぼ半減。
// BenchmarkSearchDepth3-20           1402            862434 ns/op    8048 allocs/op
// StageJ: 静止探索(quiesce)を常時有効化・killer/history手順序を追加。
// killer/historyは通常探索のノードを大きく減らすが、quiesceが末端で
// 追加のノードを掘るためnegamax全体では正味増加(質と引き換えの想定内の
// 悪化。TestQuiesceAvoidsHorizonEffectBlunder等で正しさの向上を別途検証)。
// BenchmarkSearchDepth3-20            336           3562927 ns/op   48747 allocs/op
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

// Stage K-2: Lazy SMP は本質的に非決定的(共有TT経由でworker間の情報が
// 混ざるため、探索順序次第で結果が変わりうる)。ここでは緩い検証として、
// (a) 返る手が合法手集合に含まれること、(b) スコアが単独探索の結果と
// 大きく乖離しない(歩1枚=100点程度)ことだけを確認する。
// 決定的な比較が必要な既存テストは引き続き Parallel=false のまま。
func TestParallelResultIsReasonablyCloseToSingleThreaded(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	depth := 4

	single, err := search.Best(b, search.Options{Depth: depth, TT: true})
	if err != nil {
		t.Fatalf("Best() (single) error: %v", err)
	}

	parallel, err := search.Best(b, search.Options{Depth: depth, Parallel: true, Workers: 4, TT: true})
	if err != nil {
		t.Fatalf("Best() (parallel) error: %v", err)
	}

	if parallel.Action == nil {
		t.Fatalf("expected a parallel action, got nil")
	}

	legal := b.Candidate()
	found := false
	for _, a := range legal {
		if a.String() == parallel.Action.String() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("parallel result %s is not a legal move: %v", parallel.Action.String(), legal)
	}

	diff := single.Score - parallel.Score
	if diff < 0 {
		diff = -diff
	}
	const pawnValue = 100
	if diff > pawnValue {
		t.Errorf("expected parallel score (%d) to be close to single-threaded score (%d), diff=%d exceeds a pawn's worth",
			parallel.Score, single.Score, diff)
	}
}

// BenchmarkSearchDepth4TT: depth4・置換表ありの速度計測。
// StageD: 17309642 ns/op
// StageI: DoMove/UndoMove化。6094286 ns/op
// StageJ: 静止探索+killer/history追加。ノード自体はkiller/historyで
// 大幅減(TestKillerHistoryReducesNodesAtDepth4参照)だが、quiesceの
// 追加探索コストが上回り正味では悪化。
// BenchmarkSearchDepth4TT-20           46          25469946 ns/op
// StageK: 置換表のロックフリー化(この関数自体はParallel=falseなので
// 直接の影響は無いが計測値を記録)。
// BenchmarkSearchDepth4TT-20           46          24900774 ns/op
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

// BenchmarkSearchDepth4Parallel: Stage K の Lazy SMP(depth4, TTあり,
// Workers自動)。StageD で報告された「Parallel+TTでGB級」のメモリ問題が
// 解消され、共有TT1枚(~24MB相当)+worker毎の探索状態のみのオーダーに
// なっていることを allocs/バイト数で確認する(単独探索と同オーダーで
// あるべき)。
// 実測: 16153885 ns/op(単独24900774 ns/opより高速)、47569819 B/op
// (単独29000279 B/opの約1.6倍。GB級だったStageDの問題は解消)。
func BenchmarkSearchDepth4Parallel(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		search.Best(board, search.Options{Depth: 4, Parallel: true, TT: true})
	}
}

// BenchmarkSearchDepth5TT: depth5・単独探索(置換表あり)の速度計測。
// Lazy SMP の実効速度向上を比較するための基準値。
// 実測: 87459275 ns/op(約87.5ms)
func BenchmarkSearchDepth5TT(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		search.Best(board, search.Options{Depth: 5, Parallel: false, TT: true})
	}
}

// BenchmarkSearchDepth5Parallel: depth5・Lazy SMP(置換表あり、
// Workers自動)の速度計測。同一深さでの wall time が単独探索(
// BenchmarkSearchDepth5TT)より短縮されることが目標(浅い探索では並列
// オーバーヘッドが勝つ場合もあるため、未達なら数値をそのまま報告する)。
// 実測: 55375067 ns/op(約55.4ms)。単独(約87.5ms)比で約37%短縮、
// 目標達成。
func BenchmarkSearchDepth5Parallel(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		search.Best(board, search.Options{Depth: 5, Parallel: true, TT: true})
	}
}

// negamax の冒頭にある千日手チェックが機能し、RepetitionDraw の局面で
// 0 を返すことを単体で確認する。
func TestNegamaxReturnsZeroForRepetitionDraw(t *testing.T) {

	b, err := shogi.NewBoard("sfen k8/9/9/9/4K4/9/9/9/9 b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	//双方の玉が2マスを往復するだけの手順を3周し、4回目の同一局面(Draw)に到達させる。
	cycle := []string{"5e5d", "9a9b", "5d5e", "9b9a"}
	for cy := 0; cy < 3; cy++ {
		for _, mov := range cycle {
			if !b.Action(shogi.NewAction(mov)) {
				t.Fatalf("Action(%s) failed", mov)
			}
		}
	}

	if b.Repetition() != shogi.RepetitionDraw {
		t.Fatalf("setup error: expected RepetitionDraw, got %v", b.Repetition())
	}

	var nodes int64
	score := search.ExportNegamax(context.Background(), b, 2, 1, -search.MateScore-1, search.MateScore+1, &nodes, nil)
	if score != 0 {
		t.Errorf("expected negamax to return 0 for a repetition-draw position, got %d", score)
	}
}

// BenchmarkSearchDepth4NoTT: depth4・置換表なしの速度計測。
// StageD: 23967738 ns/op
// StageI: DoMove/UndoMove化。7719995 ns/op
// StageJ: 静止探索+killer/history追加。
// BenchmarkSearchDepth4NoTT-20         51          25926202 ns/op
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
