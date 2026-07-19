package shogi

import (
	"fmt"
	"strings"
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

// 動作文字列から動作を作成
func NewAction(buf string) *Action {

	var m Action

	m.before = PosNone
	m.after = PosNone
	m.hit = 0
	m.growth = false

	if buf == Resign {
		m.resign = true
		return &m
	}

	leng := len(buf)
	if leng == 2 {
		//元位置のみ
		m.before = parsePos(buf)
	} else if leng >= 4 {
		//打ちの場合
		if strings.Index(buf, "*") == 1 {
			m.hit = buf[0]
			m.after = parsePos(buf[2:4])
		} else {
			//ならない移動
			m.before = parsePos(buf[0:2])
			m.after = parsePos(buf[2:4])
			if leng == 5 {
				//なる移動
				m.growth = true
			}
		}
	} else {
		return nil
	}

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

// 動作文字列に変換
func (m Action) String() string {

	if m.resign {
		return Resign
	}

	if m.Hit() {
		return string(m.hit) + "*" + m.after.String()
	}
	g := ""
	if m.growth {
		g = "+"
	}
	return m.before.String() + m.after.String() + g
}

// デバッグ用の文字列
func (m Action) GoString() string {
	return fmt.Sprintf("%v -> %v", m.before, m.after)
}
