package shogi

import (
	"fmt"
	"log/slog"
	"strings"
)

type TurnType string

const (
	TurnBlack TurnType = "b"
	TurnWhite TurnType = "w"
)

func (t TurnType) Next() TurnType {
	if t == TurnBlack {
		return TurnWhite
	}
	return TurnBlack
}

func (t TurnType) Index() int {
	if t == TurnBlack {
		return 0
	}
	return 1
}

// 盤面
type Board struct {
	startSFEN string
	turn      TurnType

	camps [2]*CampBoard
}

const (
	SFENLine = "/"
)

func NewBoard(sfen string) (*Board, error) {

	var b Board
	b.startSFEN = sfen

	b.turn = TurnBlack

	b.camps[0] = NewCampBoard(TurnBlack)
	b.camps[1] = NewCampBoard(TurnWhite)
	//相手を設定
	b.camps[0].setOpposite(b.camps[1])

	err := b.parseSFEN()
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (b *Board) Copy() *Board {

	var dst Board
	dst.startSFEN = b.startSFEN
	dst.turn = b.turn
	dst.camps[0] = b.camps[0].copy()
	dst.camps[1] = b.camps[1].copy()
	dst.camps[0].setOpposite(dst.camps[1])

	return &dst
}

// 座標位置にピースを配置
func (b *Board) set(x, y int, piece string) error {
	p := NewPieceFromString(piece)
	p.set(x, y)
	b.camps[p.turn.Index()].set(x, y, p.typ)
	return nil
}

func (b *Board) parseSFEN() error {

	s := strings.Split(b.startSFEN, SFENLine)
	if len(s) != 9 {
		return fmt.Errorf("SFEN parse error")
	}

	//TODO 成の考慮がない

	for y := 1; y <= 9; y++ {

		line := s[y-1]
		x := 0

		for idx := 0; idx < len(line); idx++ {
			c := line[idx]
			if c >= '0' && c <= '9' {
				x = x + (int(c) - 48)
			} else {
				b.set(x+1, y, string(c))
				x++
			}
		}
		if x != 9 {
			return fmt.Errorf("SFEN parse error")
		}
	}
	return nil
}

func (b *Board) setTurn(t TurnType) {
	b.turn = t
}

func (b *Board) setHave(buf string) {
	//TODO 初期設定で駒を持たせる
	slog.Error(fmt.Sprintf("Not Implemented"))
}

// 動作させる
func (b *Board) Action(a *Action) bool {

	rtn := false
	rtn = b.camps[b.turn.Index()].action(a)
	if !rtn {
		return false
	}
	b.turn = b.turn.Next()

	return true
}

// 動かせる箇所を取得
func (b *Board) Can(check bool) []*Action {
	now := b.camps[b.turn.Index()]
	actions := now.can()
	if !check {
		return actions
	}

	return b.filter(actions)
}

// 次の手の後に王手がある場合、阻止する手のみ羅列
func (b *Board) filter(actions []*Action) []*Action {

	var newActions []*Action

	slog.Info(fmt.Sprintf("can:[%d]", len(actions)))
	//全操作を設定
	for _, a := range actions {

		n := b.Copy()
		b.Action(a)

		//次の手番の可能性のある操作をすべて取得
		n_actions := n.Can(false)
		check := false

		for _, na := range n_actions {
			if na.enemy != nil {
				//取れてる
				if na.enemy.typ == King {
					check = true
					break
				}
			}
		}

		if !check {
			newActions = append(newActions, a)
		}
	}

	slog.Info(fmt.Sprintf("after can:[%d]", len(newActions)))
	return newActions
}

// 座標にある駒を取得
func (b *Board) Get(x, y int) *Piece {
	p := b.camps[0].Get(x, y)
	if !p.Empty() {
		return p
	}
	return b.camps[1].Get(x, y)
}

// 盤面表示
func (b *Board) GoString() string {

	var builder strings.Builder

	black := "*"
	white := " "
	if b.turn == TurnWhite {
		black = " "
		white = "*"
	}

	builder.WriteString(fmt.Sprintf(" |  1  2  3  4  5  6  7  8  9  |\n"))
	builder.WriteString(fmt.Sprintf("-------------------------------|\n"))

	for y := 1; y <= 9; y++ {
		builder.WriteString(fmt.Sprintf("%d|  ", y))
		for x := 1; x <= 9; x++ {
			p := b.Get(x, y)
			builder.WriteString(fmt.Sprintf("%-3s", p.Mark()))
		}
		builder.WriteString("|\n")
	}
	builder.WriteString(fmt.Sprintf("-------------------------------|\n"))
	builder.WriteString(fmt.Sprintf("%sBlack:%v\n", black, b.camps[0].has))
	builder.WriteString(fmt.Sprintf("%sWhite:%v\n", white, b.camps[1].has))
	return builder.String()
}
