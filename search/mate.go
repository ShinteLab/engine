package search

import (
	"context"

	"shogi"
)

// ctx.Err() を確認するノード数間隔(negamaxと同じ間隔を使う)
const mateCancelCheckInterval = cancelCheckInterval

// Mate は現手番側が maxDepth 手以内(奇数手)で相手玉を詰ませられるかを
// 素朴な AND-OR 深さ優先探索(反復深化)で調べる。
// df-pn 等の高度なアルゴリズムは過剰実装のため採用しない。
// 無駄合い(意味の無い合駒)の除外は実装しない(スコープ外)。
//
// 詰む場合は攻方・玉方交互の手順を返す(found=true)。
// 詰まない、または maxDepth 手以内に詰みが見つからない場合は found=false。
// ctx がキャンセルされた場合は found=false, err=ctx.Err() を返す
// (この場合、結果は「詰みなし」を意味しない。呼び出し側で区別すること)。
func Mate(ctx context.Context, b *shogi.Board, maxDepth int) ([]*shogi.Action, bool, error) {

	if maxDepth <= 0 {
		maxDepth = 1
	}

	var nodes int64

	//1, 3, 5, ... と手数を伸ばしながら最短の詰みを探す
	for depth := 1; depth <= maxDepth; depth += 2 {

		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}

		moves, ok := mateOr(ctx, b, depth, &nodes)

		if ctx.Err() != nil {
			//探索途中で打ち切られた場合、oksの結果は信頼できないため
			//「詰みなし」ではなくエラーとして扱う
			return nil, false, ctx.Err()
		}

		if ok {
			return moves, true, nil
		}
	}

	return nil, false, nil
}

// 攻方ノード(OR): 王手になる手のいずれかで詰みに至れば成功。
// depth は残り手数(攻方の手を含む)。
func mateOr(ctx context.Context, b *shogi.Board, depth int, nodes *int64) ([]*shogi.Action, bool) {

	*nodes++
	if *nodes%mateCancelCheckInterval == 0 && ctx.Err() != nil {
		return nil, false
	}

	if depth <= 0 {
		return nil, false
	}

	actions := b.Candidate()
	for _, a := range actions {

		u, ok := b.DoMove(a)
		if !ok {
			continue
		}

		//王手になる手のみを試す(b.Turn()は玉方の手番)
		if !b.IsCheck(b.Turn()) {
			b.UndoMove(u)
			continue
		}

		defActions := b.Candidate()
		if len(defActions) == 0 {
			//応手が無い = この手で即詰み
			b.UndoMove(u)
			return []*shogi.Action{a}, true
		}

		if depth == 1 {
			//王手はしたが即詰みではなく、これ以上深さの余裕が無い
			b.UndoMove(u)
			continue
		}

		subMoves, ok2 := mateAnd(ctx, b, depth-1, nodes)
		b.UndoMove(u)
		if ok2 {
			moves := make([]*shogi.Action, 0, len(subMoves)+1)
			moves = append(moves, a)
			moves = append(moves, subMoves...)
			return moves, true
		}
	}

	return nil, false
}

// 玉方ノード(AND): すべての応手が詰みに至る場合のみ詰み成立。
// 合法手が無い場合は(呼び出し元でチェック済みのはずだが)詰み扱いとする。
// depth は残り手数(玉方の手を含む)。
func mateAnd(ctx context.Context, b *shogi.Board, depth int, nodes *int64) ([]*shogi.Action, bool) {

	*nodes++
	if *nodes%mateCancelCheckInterval == 0 && ctx.Err() != nil {
		return nil, false
	}

	actions := b.Candidate()
	if len(actions) == 0 {
		return nil, true
	}

	//玉方はどの応手でも詰むので、返す手順としては最初の応手を代表として使う。
	//ただし詰みの成立判定自体は全応手を確認する必要がある。
	var pv []*shogi.Action

	for _, a := range actions {

		u, ok := b.DoMove(a)
		if !ok {
			continue
		}

		subMoves, ok2 := mateOr(ctx, b, depth-1, nodes)
		b.UndoMove(u)

		if !ok2 {
			//逃げ道が1つでもあれば詰みではない
			return nil, false
		}

		if pv == nil {
			moves := make([]*shogi.Action, 0, len(subMoves)+1)
			moves = append(moves, a)
			moves = append(moves, subMoves...)
			pv = moves
		}
	}

	return pv, true
}
