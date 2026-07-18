package shogi

import (
	"fmt"
	"log/slog"
)

// 片方の盤面
type CampBoard struct {
	board     BitBoard
	typBoards [14]BitBoard

	has Pieces

	ownTurn TurnType
	enemy   *CampBoard
}

// 片側の盤面を作成
func NewCampBoard(t TurnType) *CampBoard {
	var b CampBoard
	b.ownTurn = t
	b.has = NewPieces()
	return &b
}

func (b *CampBoard) copy() *CampBoard {
	dst := *b
	dst.enemy = nil
	return &dst
}

// 存在する駒を取得
func (b *CampBoard) Get(x, y int) *Piece {

	if !b.board.get(x, y) {
		return nil
	}

	for t := 0; t < len(b.typBoards); t++ {
		p := &b.typBoards[t]
		if p.get(x, y) {
			rtn := NewPieceFromType(PieceType(t), b.ownTurn)
			rtn.set(x, y)
			return rtn
		}
	}
	return nil
}

// 相手を設定
func (b *CampBoard) setEnemy(e *CampBoard) {
	b.enemy = e
	e.enemy = b
}

// 動作させる
func (b *CampBoard) action(m *Action) bool {

	//元位置の駒を取得してを削除
	p, err := b.setMove(m)
	if err != nil {
		slog.Error(fmt.Sprintf("setMove() error: %v", err))
		return false
	}

	//相手位置に居れば削除し、自分に追加
	f, err := b.getEnemy(p)
	if err != nil {
		slog.Error(fmt.Sprintf("getEnemy() error: %v", err))
		return false
	}

	//打ちで相手位置にいた場合
	if f && m.Hit() {
		slog.Error(fmt.Sprintf("Hit and Enemy error: %v", err))
		return false
	}

	return true
}

// 盤面に設定
func (b *CampBoard) set(x, y int, t PieceType) {
	b.typBoards[t].set(x, y)
	b.board.set(x, y)
}

// 駒を動かす
func (b *CampBoard) setMove(action *Action) (Pos, error) {

	var t PieceType
	if !action.Hit() {
		t = b.clear(action.before)
		if t == PieceTypeNotFound {
			return PosNone, fmt.Errorf("NotFound [%s]", action.before)
		}
	}

	x1, y1 := action.after.XY()
	//成りの指示があり、成ってない場合
	if action.growth && !t.IsGrowth() {
		if action.GrowthArea(b.ownTurn) {
			wkT := t.Growth()
			if wkT != PieceTypeNotFound {
				t = wkT
			} else {
				slog.Error(fmt.Sprintf("growth error[%v][%v]", t, action))
			}
		} else {
			return PosNone, fmt.Errorf("Growth error[%s]", action)
		}
	} else if action.Hit() {
		// タイプを取得
		t = action.HitType()
		//TODO 自駒から減らす
		b.has.remove(t)
	}

	b.typBoards[t].set(x1, y1)
	b.board.set(x1, y1)

	//移動位置を返す
	return action.after, nil
}

// 駒を消す
func (b *CampBoard) clear(p Pos) PieceType {

	x, y := p.XY()

	t := PieceTypeNotFound
	for idx := 0; idx < len(b.typBoards); idx++ {
		if b.typBoards[idx].clear(x, y) {
			t = PieceType(idx)
			b.board.clear(x, y)
			break
		}
	}
	return t
}

// 相手から駒を取る
func (b *CampBoard) getEnemy(p Pos) (bool, error) {
	is := false
	t := b.enemy.clear(p)
	if t != PieceTypeNotFound {
		b.has.add(t)
		is = true
	}
	return is, nil
}

// 動作できる箇所を返す(擬似合法手。王手放置・自殺手・打ち歩詰めを含みうる)
func (b *CampBoard) pseudoCandidate() []*Action {

	actions := make([]*Action, 0, 256)
	//盤面すべての動作できる場所を取得
	for idx := 0; idx < len(b.typBoards); idx++ {
		a := b.canBit(PieceType(idx))
		if a != nil {
			actions = append(actions, a...)
		}
	}

	//全盤面の空き座標に打ちを作成
	poss := b.emptyPos()
	//全持ち駒の種類を取得
	types := b.hasAll()
	for _, pos := range poss {
		x, y := pos.XY()
		sq := squareOf(x, y)
		for _, t := range types {

			p2 := false
			if t == Pawn {
				p2 = b.typBoards[0].hasX(x)
			}

			if !p2 {
				if hasMoveMask(b.ownTurn, t, sq) {
					a := NewAction(fmt.Sprintf("%s*%v", t.Mark(), pos))
					actions = append(actions, a)
				}
			}
		}
	}
	return actions
}

// 自軍全駒の利き集合を返す(王手判定用)。
// 自駒の占有マスも含めてよい(「利いているか」だけを見るため andNot しない)。
func (b *CampBoard) attackAll(occ *BitBoard) BitBoard {

	var result BitBoard
	for t := 0; t < len(b.typBoards); t++ {
		bit := &b.typBoards[t]
		if bit.isZero() {
			continue
		}
		pt := PieceType(t)
		bit.forEach(func(sq int) {
			a := attacks(b.ownTurn, pt, sq, occ)
			result.or(&a)
		})
	}
	return result
}

// 空き位置を作成して空の部分を取得
func (b *CampBoard) emptyPos() []Pos {

	bit := b.board
	bit.or(&b.enemy.board)
	//slog.Info(fmt.Sprintf("%#v", bit))
	//全盤面を作成
	//空いているビットを作成
	return bit.empty()
}

// 持っている駒の種別を取得
func (b *CampBoard) hasAll() []PieceType {
	return b.has.all()
}

// その駒での移動箇所を取得
func (b *CampBoard) canBit(t PieceType) []*Action {

	actions := make([]*Action, 0, 16)
	bit := &b.typBoards[t]

	if bit.isZero() {
		return actions
	}

	//すでに成っている
	g := t.IsGrowth()

	//走り駒のブロッカー判定に使う全体の占有盤面
	occ := b.board
	occ.or(&b.enemy.board)

	bit.forEach(func(sq int) {

		x, y := sqXY(sq)
		before := newPos(x, y)

		//利きを取得し、自陣の駒がある場所を除外
		moves := attacks(b.ownTurn, t, sq, &occ)
		moves.andNot(&b.board)

		moves.forEach(func(dst int) {

			dx, dy := sqXY(dst)
			np := newPos(dx, dy)

			//敵を検索
			var e *Piece
			if b.enemy.board.getSq(dst) {
				e = b.enemy.Get(dx, dy)
			}

			mv := newMoveAction(before, np, false)

			//まだ成ってないで成る場所の場合
			if !g && t != King && t != Gold && mv.GrowthArea(b.ownTurn) {

				//成を追加
				gmv := newMoveAction(before, np, true)
				gmv.SetEnemy(e)
				actions = append(actions, gmv)

				//次に動く場所がない場合(行き所のない駒)は不成を出さない
				if !hasMoveMask(b.ownTurn, t, dst) {
					mv = nil
				}
			}

			if mv != nil {
				mv.SetEnemy(e)
				actions = append(actions, mv)
			}
		})
	})
	return actions
}
