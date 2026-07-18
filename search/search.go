// Package search は shogi パッケージの Board を対象にした
// Negamax + αβ 枝刈り + 反復深化による簡易な指し手探索を提供する。
package search

import (
	"context"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"shogi"
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
	Parallel bool // ルート並列探索を行うか
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
// opt.TT が true の場合、置換表を使う。ルート並列時(opt.Parallel)は
// ルート goroutine ごとに独立した置換表インスタンスを持たせる
// (共有すると data race になるため。共有 TT + ロックフリーは将来課題)。
// 単一スレッド時は反復深化の各深さを通じて1つの置換表を使い回す。
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

	var tt *transTable
	if opt.TT && !opt.Parallel {
		tt = newTransTable()
	}

	var final Result
	for d := 1; d <= depth; d++ {

		//深さ1は(呼び出し元がすでに期限切れのctxを渡していても)必ず
		//試みる。それ以外は開始前にキャンセルを確認して打ち切る。
		if d > 1 && ctx.Err() != nil {
			break
		}

		var res Result
		if opt.Parallel {
			res = searchRootParallel(ctx, b, actions, d, opt.TT)
		} else {
			res = searchRoot(ctx, b, actions, d, tt)
		}

		if res.Action != nil {
			final = res
			actions = reorderWithBest(actions, res.Action)

			if opt.Info != nil {
				opt.Info(Info{
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

// 単一スレッドでのルート探索
func searchRoot(ctx context.Context, b *shogi.Board, actions []*shogi.Action, depth int, tt *transTable) Result {

	var nodes int64
	alpha := -infScore
	beta := infScore

	var bestAction *shogi.Action
	bestScore := -infScore

	for _, a := range actions {

		if ctx.Err() != nil {
			break
		}

		nb := b.Copy()
		if !nb.Action(a) {
			continue
		}

		score := -negamax(ctx, nb, depth-1, 1, -beta, -alpha, &nodes, tt)

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

// ルート並列探索。ルートの手数分だけ goroutine を起こし、各手は独立した
// αβ窓(-∞,+∞)で探索する(共有 alpha は使わずシンプルさを優先する)。
// useTT が true の場合、各 goroutine は自前の置換表インスタンスを持つ
// (goroutine 間で共有しないため data race にならない)。
// ノード数のみ atomic に集計する。
func searchRootParallel(ctx context.Context, b *shogi.Board, actions []*shogi.Action, depth int, useTT bool) Result {

	n := len(actions)
	scores := make([]int, n)
	oks := make([]bool, n)
	var nodes int64

	var wg sync.WaitGroup
	wg.Add(n)
	for i, a := range actions {
		go func(i int, a *shogi.Action) {
			defer wg.Done()

			if ctx.Err() != nil {
				return
			}

			nb := b.Copy()
			if !nb.Action(a) {
				return
			}

			var workerTT *transTable
			if useTT {
				workerTT = newTransTable()
			}

			var localNodes int64
			score := -negamax(ctx, nb, depth-1, 1, -infScore, infScore, &localNodes, workerTT)
			atomic.AddInt64(&nodes, localNodes)

			if ctx.Err() != nil {
				//打ち切りによる途中結果は信頼できないため採用しない
				return
			}

			scores[i] = score
			oks[i] = true
		}(i, a)
	}
	wg.Wait()

	bestIdx := -1
	bestScore := -infScore
	for i := 0; i < n; i++ {
		if !oks[i] {
			continue
		}
		if bestIdx == -1 || scores[i] > bestScore {
			bestScore = scores[i]
			bestIdx = i
		}
	}

	if bestIdx == -1 {
		return Result{Nodes: nodes}
	}

	return Result{Action: actions[bestIdx], Score: bestScore, Nodes: nodes}
}

// Negamax + αβ枝刈り本体。
// depth は残り探索深さ、ply はルートからの手数(詰みスコアの補正に使う)。
// tt が nil の場合は置換表を使わない。
func negamax(ctx context.Context, b *shogi.Board, depth, ply int, alpha, beta int, nodes *int64, tt *transTable) int {

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

	if tt != nil {
		if score, best, ok := tt.probe(b.Hash(), depth, ply, alpha, beta); ok {
			return score
		} else {
			ttBest = best
		}
	}

	if depth == 0 {
		return eval(b)
	}

	actions := b.Candidate()
	if len(actions) == 0 {
		//合法手が無い場合は詰み(王手されていない場合もステイルメイト相当として
		//同じ負け値を返す)。ply が小さい(=早い詰み)ほど評価値の絶対値が
		//大きくなるようにして、早い詰みを優先する。
		return -(MateScore - ply)
	}

	orderMoves(actions, ttBest)

	best := -infScore
	var bestMove *shogi.Action
	for _, a := range actions {
		nb := b.Copy()
		if !nb.Action(a) {
			continue
		}

		score := -negamax(ctx, nb, depth-1, ply+1, -beta, -alpha, nodes, tt)

		if score > best {
			best = score
			bestMove = a
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			//βカット
			break
		}

		if ctx.Err() != nil {
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

// 手番側から見た「自分の駒価値の合計 - 相手の駒価値の合計」
func eval(b *shogi.Board) int {
	t := b.Turn()
	return b.Material(t) - b.Material(t.Next())
}

// 手の並べ替え: 置換表の最善手(あれば)を最優先し、
// 次に捕獲手(取る駒の価値降順)を先頭に。
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
