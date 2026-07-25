package shogi

import (
	"fmt"

	"github.com/ShinteLab/core/usi"
)

type Actions []*Action

func (a Actions) String() string {
	return fmt.Sprintf("%v", []*Action(a))
}

// 動作
type Action struct {
	resign bool

	before Pos
	after  Pos
	growth bool

	hit byte
	//動作予想時に相手がafterに居たらいれる
	enemy *Piece
	//敵の次の１手
	next *Action
}

func ResignAction() *Action {
	return NewAction(Resign)
}

// 移動から動作を作成(候補手生成用)
func newMoveAction(before, after Pos, growth bool) *Action {
	return &Action{before: before, after: after, growth: growth}
}

// 動作文字列から動作を作成。手表記の仕様は core/usi に集約している。
func NewAction(buf string) *Action {

	mv, ok := usi.ParseMove(buf)
	if !ok {
		return nil
	}

	var m Action
	if mv.Resign {
		m.resign = true
		m.before = PosNone
		m.after = PosNone
		return &m
	}

	//打ちでない場合 mv.Drop==0、移動元/先が無い成分は 0 のため PosNone になる
	m.hit = mv.Drop
	m.before = newPos(mv.FromX, mv.FromY)
	m.after = newPos(mv.ToX, mv.ToY)
	m.growth = mv.Promote
	return &m
}

func (a *Action) Hit() bool {
	return a.hit != 0
}

func (a *Action) HitType() PieceType {
	p, _ := NewPieceType(a.hit)
	return p
}

func (a *Action) Enemy() *Piece {
	return a.enemy
}

// 移動元の座標(打ちの場合は未定義。Hit()で確認すること)
func (a *Action) BeforeXY() (int, int) {
	return a.before.XY()
}

// 移動先(打ち先)の座標
func (a *Action) AfterXY() (int, int) {
	return a.after.XY()
}

// 動作時に取った駒を設定
func (a *Action) SetEnemy(p *Piece) {
	a.enemy = p
}

// この Action が成りを指定しているか(打ちの場合は常にfalse)
func (a *Action) Promotes() bool {
	return a.growth
}

// 成っていい場所かを判定
// 成れる駒のみ処理を行う
func (m *Action) GrowthArea(turn TurnType) bool {

	if m.Hit() {
		return false
	}

	_, y0 := m.before.XY()
	_, y1 := m.after.XY()

	return turn.GrowthArea(y0, y1)
}

func (m *Action) SetNext(a *Action) {
	m.next = a
}

func (m *Action) Next() *Action {
	return m.next
}

// 動作文字列に変換。手表記の仕様は core/usi に集約している。
func (m Action) String() string {

	if m.resign {
		return usi.Resign
	}

	mv := usi.Move{Drop: m.hit, Promote: m.growth}
	mv.FromX, mv.FromY = m.before.XY()
	mv.ToX, mv.ToY = m.after.XY()
	return mv.String()
}

// デバッグ用の文字列
func (m Action) GoString() string {
	return fmt.Sprintf("%v -> %v", m.before, m.after)
}
