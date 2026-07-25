// Package search は shogi パッケージの Board を対象にした
// Negamax + αβ 枝刈り + 反復深化による簡易な指し手探索を提供する。
package search

import (
	"context"
	"errors"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"shinte/engine"
)

// 反復深化の各深さ完了時に呼ばれる情報コールバック向けの構造体。
type Info struct {
	Depth   int
	ScoreCP int
	Nodes   int64
	PV      []*shogi.Action
	Mate    int
}

// 探索オプション
type Options struct {
	Depth    int  // 探索深さ(0以下の場合は3を使う)
	Parallel bool // Lazy SMP(複数workerが共有TTを使って同じ局面を探索)を行うか
	Workers  int  // Parallel時のworker数(0以下ならruntime.NumCPU()と8の小さい方)
	TT       bool // 置換表を使うか(ゼロ値=false で無効)

	Movetime time.Duration // 0=無制限。指定時は BestContext 内で ctx に deadline を重ねる
	Info     func(Info)    // 反復深化の各深さ完了時に呼ばれる(nilなら未使用)
}

// 探索結果
type Result struct {
	Action *shogi.Action
	Score  int
	Nodes  int64 // 訪問ノード数
}

// 合法手が1つも無い(詰み/ステイルメイト相当)場合に返すエラー
var ErrNoMoves = errors.New("search: no legal moves")

// Material の最大値より十分大きい詰みスコアの基準値
const MateScore = 1 << 20

// 内部でのαβ初期窓の上下限(詰みスコアより広く取る)
const infScore = MateScore + 1

// ctx.Err() を確認するノード数間隔
const cancelCheckInterval = 1024

// killer手を記録するply(ルートからの手数)の上限。
// これを超えるplyでは記録・参照をスキップする(安全側に倒す)。
const maxKillerPly = 128

// 1回の探索インスタンス(BestContext呼び出し1回、またはルート並列時は
// goroutine 1つ)が持つ可変状態。置換表・killer手・historyテーブルは
// 反復深化の各深さを通じて使い回すが、並列ルート間では共有しない
// (goroutineごとに newSearcher する)。
type searcher struct {
	tt *transTable

	//ply毎の killer 手(beta カットを起こした捕獲でも ttBest でもない手)。
	//同一手の重複記録はしない。
	killer [maxKillerPly][2]*shogi.Action

	//手番Index × 移動元sq × 移動先sq のhistoryスコア。打ちは対象外
	//(シンプルさ優先。打ちのhistoryは今回のスコープ外)。
	history [2][81][81]int
}

func newSearcher(tt *transTable) *searcher {
	return &searcher{tt: tt}
}

// Best は BestContext(context.Background(), b, opt) の薄いラッパ。
func Best(b *shogi.Board, opt Options) (Result, error) {
	return BestContext(context.Background(), b, opt)
}

// BestContext は現局面の最善手を Negamax + αβ + 反復深化で探索して返す。
// ctx がキャンセルされた場合、探索を打ち切って「完了した最後の深さの結果」
// (それすら無ければ先頭手)を返す。必ず有効な Action を返す(合法手が
// 1つも無い場合を除く)。
//
// opt.Movetime > 0 の場合、ctx に対して deadline を重ねる。
// opt.TT が true の場合、置換表を使う。
// opt.Parallel が true の場合、Lazy SMP(opt.Workers 個の worker が
// 同一の共有置換表を使って同じルート局面を独立に反復深化探索する。
// worker0 が主で、その結果・Info コールバックのみを採用する)で探索する。
// 静止探索(quiesce)は常時有効(depth==0でevalの代わりに呼ぶ)。
func BestContext(ctx context.Context, b *shogi.Board, opt Options) (Result, error) {

	depth := opt.Depth
	if depth <= 0 {
		depth = 3
	}

	if opt.Movetime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Movetime)
		defer cancel()
	}

	actions := b.Candidate()
	if len(actions) == 0 {
		return Result{}, ErrNoMoves
	}

	orderMoves(actions, nil)

	if opt.Parallel {
		return bestLazySMP(ctx, b, actions, depth, opt)
	}

	var tt *transTable
	if opt.TT {
		tt = newTransTable()
	}
	s := newSearcher(tt)

	return iterativeDeepen(ctx, b, actions, depth, s, opt.Info)
}

// 単一スレッドの反復深化ループ本体。Parallel=false の BestContext と、
// Lazy SMP の worker0 で共用する。
func iterativeDeepen(ctx context.Context, b *shogi.Board, actions []*shogi.Action, maxDepth int, s *searcher, info func(Info)) (Result, error) {

	var final Result
	for d := 1; d <= maxDepth; d++ {

		//深さ1は(呼び出し元がすでに期限切れのctxを渡していても)必ず
		//試みる。それ以外は開始前にキャンセルを確認して打ち切る。
		if d > 1 && ctx.Err() != nil {
			break
		}

		res := searchRoot(ctx, b, actions, d, s)

		if res.Action != nil {
			final = res
			actions = reorderWithBest(actions, res.Action)

			if info != nil {
				info(Info{
					Depth:   d,
					ScoreCP: res.Score,
					Nodes:   res.Nodes,
					PV:      []*shogi.Action{res.Action},
				})
			}
		}
	}

	if final.Action == nil {
		//深さ1すら完走できなかった場合(ctxが最初から期限切れ等)でも
		//必ず有効な手を返す: 先頭手(捕獲優先で並べ替え済み)を返す。
		final.Action = actions[0]
	}

	return final, nil
}

// Lazy SMP: opt.Workers 個の worker(0なら runtime.NumCPU() と 8 の
// 小さい方)が、同一の共有置換表を使って同じルート局面を独立に
// 反復深化探索する。各 worker はルート局面を1回だけ Copy() して
// 専用インスタンスを持ち(盤面は共有しない)、DoMove/UndoMove で
// 掘り下げる。共有するのは置換表のみで、killer/history は
// worker ごとに独立した searcher が持つ(共有すると data race になる
// ため)。
//
// worker0 が主で、各深さを完了するたびにその結果を採用し Info を発火する。
// 他 worker は開始深さを +1 ずつずらして探索することで置換表の
// カバー範囲を広げる(標準的な Lazy SMP)。worker0 が目標深さを完了したら
// 全 worker をまとめてキャンセルする。
func bestLazySMP(ctx context.Context, b *shogi.Board, actions []*shogi.Action, depth int, opt Options) (Result, error) {

	workers := opt.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
		if workers > 8 {
			workers = 8
		}
	}
	if workers < 1 {
		workers = 1
	}

	var tt *transTable
	if opt.TT {
		tt = newTransTable()
	}

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	var mu sync.Mutex
	var final Result
	haveResult := false

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(w int) {
			defer wg.Done()

			nb := b.Copy()
			s := newSearcher(tt)

			localActions := make([]*shogi.Action, len(actions))
			copy(localActions, actions)

			//worker0は1から、他workerは開始深さを+1ずつずらす
			startDepth := 1 + w

			for d := startDepth; d <= depth; d++ {

				if d > startDepth && workerCtx.Err() != nil {
					break
				}

				res := searchRoot(workerCtx, nb, localActions, d, s)
				if res.Action == nil {
					continue
				}
				localActions = reorderWithBest(localActions, res.Action)

				if w == 0 {
					mu.Lock()
					final = res
					haveResult = true
					mu.Unlock()

					if opt.Info != nil {
						opt.Info(Info{
							Depth:   d,
							ScoreCP: res.Score,
							Nodes:   res.Nodes,
							PV:      []*shogi.Action{res.Action},
						})
					}

					if d == depth {
						//主workerが目標深さを完了したので全workerを打ち切る
						cancelWorkers()
					}
				}
			}
		}(w)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if !haveResult || final.Action == nil {
		//主workerが1手も確定できなかった場合(ctxが最初から期限切れ等)
		//でも必ず有効な手を返す。
		return Result{Action: actions[0]}, nil
	}

	return final, nil
}

// 単一スレッドでのルート探索
func searchRoot(ctx context.Context, b *shogi.Board, actions []*shogi.Action, depth int, s *searcher) Result {

	var nodes int64
	alpha := -infScore
	beta := infScore

	var bestAction *shogi.Action
	bestScore := -infScore

	for _, a := range actions {

		if ctx.Err() != nil {
			break
		}

		u, ok := b.DoMove(a)
		if !ok {
			continue
		}

		score := -s.negamax(ctx, b, depth-1, 1, -beta, -alpha, &nodes)

		b.UndoMove(u)

		if ctx.Err() != nil {
			//この手の探索はキャンセルにより途中で打ち切られた可能性があり、
			//スコアの信頼性が無いので採用しない。
			break
		}

		if bestAction == nil || score > bestScore {
			bestScore = score
			bestAction = a
		}
		if score > alpha {
			alpha = score
		}
	}

	return Result{Action: bestAction, Score: bestScore, Nodes: nodes}
}

// Negamax + αβ枝刈り本体。
// depth は残り探索深さ、ply はルートからの手数(詰みスコアの補正に使う)。
// s.tt が nil の場合は置換表を使わない。
func (s *searcher) negamax(ctx context.Context, b *shogi.Board, depth, ply int, alpha, beta int, nodes *int64) int {

	n := atomic.AddInt64(nodes, 1)
	if n%cancelCheckInterval == 0 && ctx.Err() != nil {
		//打ち切り: 値に意味は無いが、呼び出し元がctx.Err()を見て
		//この結果を破棄する前提で即座に返す。
		return 0
	}

	//千日手判定(TTより先に見る: 千日手のスコアはこの局面固有の性質であり、
	//深さに依らないため置換表を経由する必要が無い)。
	switch b.Repetition() {
	case shogi.RepetitionDraw:
		return 0
	case shogi.RepetitionPerpetualLose:
		//現局面の手番側が王手を掛け続けた側で、その側の負け。
		//mate と同様、ply が小さい(=早く負けが確定する)ほど絶対値が大きくなる。
		return -(MateScore - ply)
	}

	origAlpha := alpha
	var ttBest *shogi.Action

	if s.tt != nil {
		if score, best, ok := s.tt.probe(b.Hash(), depth, ply, alpha, beta); ok {
			return score
		} else {
			ttBest = best
		}
	}

	if depth == 0 {
		//静止探索(常時有効): 水平線効果を防ぐため捕獲の応酬が
		//収まるまで追加で読む。
		return s.quiesce(ctx, b, alpha, beta, ply, 0, nodes)
	}

	actions := b.Candidate()
	if len(actions) == 0 {
		//合法手が無い場合は詰み(王手されていない場合もステイルメイト相当として
		//同じ負け値を返す)。ply が小さい(=早い詰み)ほど評価値の絶対値が
		//大きくなるようにして、早い詰みを優先する。
		return -(MateScore - ply)
	}

	moverIdx := b.Turn().Index()
	s.orderMovesWithHeuristics(actions, ttBest, ply, moverIdx)

	best := -infScore
	var bestMove *shogi.Action
	for _, a := range actions {
		u, ok := b.DoMove(a)
		if !ok {
			continue
		}

		score := -s.negamax(ctx, b, depth-1, ply+1, -beta, -alpha, nodes)

		b.UndoMove(u)

		if score > best {
			best = score
			bestMove = a
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			//βカット: 捕獲でも打ちでもない手なら killer/history を更新する。
			if !a.Hit() && a.Enemy() == nil {
				s.recordKiller(ply, a)
				s.recordHistory(moverIdx, a, depth)
			}
			break
		}

		if ctx.Err() != nil {
			break
		}
	}

	if s.tt != nil {
		flag := ttExact
		if best <= origAlpha {
			flag = ttUpper
		} else if best >= beta {
			flag = ttLower
		}
		s.tt.store(b.Hash(), depth, ply, best, flag, bestMove)
	}

	return best
}

// ply における killer 手を記録する(先頭2枠、同一手の重複記録はしない)。
func (s *searcher) recordKiller(ply int, a *shogi.Action) {
	if ply < 0 || ply >= maxKillerPly {
		return
	}
	if s.killer[ply][0] != nil && s.killer[ply][0].String() == a.String() {
		return
	}
	s.killer[ply][1] = s.killer[ply][0]
	s.killer[ply][0] = a
}

// history テーブルを更新する(beta カット時、depth*depth を加算)。
// 打ちは対象外(シンプルさ優先)。
func (s *searcher) recordHistory(moverIdx int, a *shogi.Action, depth int) {
	fx, fy := a.BeforeXY()
	tx, ty := a.AfterXY()
	from := squareIndex(fx, fy)
	to := squareIndex(tx, ty)
	s.history[moverIdx][from][to] += depth * depth
}

func squareIndex(x, y int) int {
	return (y-1)*9 + (x - 1)
}

// 手番側から見た「自分の駒価値の合計 - 相手の駒価値の合計」+ PST。
// (eval() 本体は search/eval.go)

// 手の並べ替え(ルート用・置換表無し): 捕獲手(取る駒の価値降順)を先頭に。
// sort.SliceStable を使い、それ以外の順序は維持する。
func orderMoves(actions []*shogi.Action, ttBest *shogi.Action) {
	sort.SliceStable(actions, func(i, j int) bool {
		return moveOrderValue(actions[i], ttBest) > moveOrderValue(actions[j], ttBest)
	})
}

const ttBestOrderValue = 1 << 30

func moveOrderValue(a, ttBest *shogi.Action) int {
	if ttBest != nil && a.String() == ttBest.String() {
		return ttBestOrderValue
	}
	return captureValue(a)
}

func captureValue(a *shogi.Action) int {
	e := a.Enemy()
	if e == nil {
		return -1
	}
	return e.Type().Value()
}

// negamax内の手の並べ替え: ttBest > 捕獲(価値降順) > killer1/killer2 >
// history降順 > その他。sort.SliceStable でそれ以外の順序は維持する。
func (s *searcher) orderMovesWithHeuristics(actions []*shogi.Action, ttBest *shogi.Action, ply, moverIdx int) {
	sort.SliceStable(actions, func(i, j int) bool {
		return s.heuristicOrderValue(actions[i], ttBest, ply, moverIdx) >
			s.heuristicOrderValue(actions[j], ttBest, ply, moverIdx)
	})
}

const (
	captureOrderBase = 1 << 20
	killer1Value     = 1 << 19
	killer2Value     = killer1Value - 1
)

func (s *searcher) heuristicOrderValue(a, ttBest *shogi.Action, ply, moverIdx int) int {

	if ttBest != nil && a.String() == ttBest.String() {
		return ttBestOrderValue
	}

	if cv := captureValue(a); cv >= 0 {
		return captureOrderBase + cv
	}

	if ply >= 0 && ply < maxKillerPly {
		if k := s.killer[ply][0]; k != nil && a.String() == k.String() {
			return killer1Value
		}
		if k := s.killer[ply][1]; k != nil && a.String() == k.String() {
			return killer2Value
		}
	}

	if !a.Hit() {
		fx, fy := a.BeforeXY()
		tx, ty := a.AfterXY()
		return s.history[moverIdx][squareIndex(fx, fy)][squareIndex(tx, ty)]
	}

	return 0
}

// 反復深化用: 前回の最善手を先頭に並べ替える(それ以外の順序は維持)。
func reorderWithBest(actions []*shogi.Action, best *shogi.Action) []*shogi.Action {

	if best == nil {
		return actions
	}

	out := make([]*shogi.Action, 0, len(actions))
	out = append(out, best)
	for _, a := range actions {
		if a == best {
			continue
		}
		out = append(out, a)
	}
	return out
}
