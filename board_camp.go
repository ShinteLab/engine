package shogi

import (
	"fmt"
	"log/slog"
)

// 片方の盤面
type CampBoard struct {
	typBoards [14]*BitBoard
	has       Pieces

	ownTurn TurnType
	oppo    *CampBoard
}

// 片側の盤面を作成
func NewCampBoard(t TurnType) *CampBoard {
	var b CampBoard
	b.ownTurn = t
	for idx := 0; idx < 14; idx++ {
		b.typBoards[idx] = NewBitBoard(PieceType(idx))
	}
	b.has = NewPieces()
	return &b
}

func (b *CampBoard) copy() *CampBoard {
	var dst CampBoard
	dst.ownTurn = b.ownTurn
	dst.has = b.has.copy()
	for idx := 0; idx < 14; idx++ {
		dst.typBoards[idx] = b.typBoards[idx].copy()
	}
	return &dst
}

// 存在する駒を取得
func (b *CampBoard) Get(x, y int) *Piece {
	for t := 0; t < len(b.typBoards); t++ {
		p := b.typBoards[t]
		if p.get(x, y) {
			rtn := NewPieceFromType(PieceType(t), b.ownTurn)
			rtn.set(x, y)
			return rtn
		}
	}
	return nil
}

// 相手を設定
func (b *CampBoard) setOpposite(oppo *CampBoard) {
	b.oppo = oppo
	oppo.oppo = b
}

// 動作させる
func (b *CampBoard) action(m *Action) bool {

	slog.Debug(fmt.Sprintf("Action[%v]", m))

	if !m.hit {
		//元位置の駒を取得してを削除
		//新しい位置に打ち込む
		p, err := b.setMove(m)
		if err != nil {
			slog.Error(fmt.Sprintf("setMove() error: %v", err))
			return false
		}

		slog.Debug(fmt.Sprintf("move[%v]", p))

		err = b.getOpposite(p)
		if err != nil {
			slog.Error(fmt.Sprintf("getOpposite() error: %v", err))
			return false
		}
	} else {

		//TODO 打ちの場合
		//誰もいないことを確認

	}
	return true
}

// 盤面に設定
func (b *CampBoard) set(x, y int, t PieceType) {
	bit := b.typBoards[t]
	bit.set(x, y)
}

// 駒を動かす
func (b *CampBoard) setMove(action *Action) (Pos, error) {

	t := b.clear(action.before)
	if t == PieceTypeNotFound {
		return PosNone, fmt.Errorf("NotFound [%s]", action.before)
	}

	x1, y1 := action.after.XY()
	if action.growth && !t.IsGrowth() {
		if action.GrowthArea(b.ownTurn) {
			t = t.Growth()
		} else {
			return PosNone, fmt.Errorf("Growth error[%s]", action)
		}
	}

	board := b.typBoards[t]
	board.set(x1, y1)

	//移動位置を返す
	return action.after, nil
}

// 駒を消す
func (b *CampBoard) clear(p Pos) PieceType {

	x, y := p.XY()

	slog.Debug(fmt.Sprintf("[%s] = %d,%d(%v)", p, x, y, b.ownTurn))
	t := PieceTypeNotFound
	for idx := 0; idx < len(b.typBoards); idx++ {
		board := b.typBoards[idx]
		if board.clear(x, y) {
			t = PieceType(idx)
			break
		}
	}
	return t
}

// 相手から駒を取る
func (b *CampBoard) getOpposite(p Pos) error {
	t := b.oppo.clear(p)
	if t != PieceTypeNotFound {
		b.has.add(t)
	}
	return nil
}

// 動作できる箇所を返す
func (b *CampBoard) can() []*Action {

	actions := make([]*Action, 0, 64)
	//盤面すべての動作できる場所を取得
	for idx := 0; idx < len(b.typBoards); idx++ {
		tb := b.typBoards[idx]
		a := b.canBit(tb)
		if a != nil {
			actions = append(actions, a...)
		}
	}

	//TODO 打ちを加算

	return actions
}

func (b *CampBoard) canBit(bit *BitBoard) []*Action {

	actions := make([]*Action, 0, 32)
	for x := 1; x <= 9; x++ {
		for y := 1; y <= 9; y++ {
			//存在する
			if bit.get(x, y) {
				//処理を行う
				a := can(bit.typ, newPos(x, y), b, b.oppo)
				if a != nil {
					actions = append(actions, a...)
				}
			}
		}
	}
	return actions
}
