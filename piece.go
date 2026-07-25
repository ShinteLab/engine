package shogi

import (
	"fmt"
	"strings"

	"github.com/ShinteLab/core/sfen"
)

// 向き等を持つ駒
type Piece struct {
	source string

	typ    PieceType
	turn   TurnType
	pos    Pos
	growth bool
}

// 文字列から駒を作成
func NewPieceFromString(p string) *Piece {

	var inst Piece
	inst.source = p
	inst.typ, inst.turn = NewPieceType(p[0])

	if len(p) >= 2 {
		inst.typ = inst.typ.Growth()
		inst.growth = true
	}
	return &inst
}

// 駒を作成
func NewPieceFromType(typ PieceType, turn TurnType) *Piece {

	var inst Piece
	inst.typ = typ
	inst.turn = turn
	inst.growth = typ.IsGrowth()
	inst.source = inst.createSource()

	return &inst
}

func (p *Piece) Type() PieceType {
	return p.typ
}

func (p *Piece) String() string {
	return fmt.Sprintf("[%v][%s]", p.pos, p.Mark())
}

// 初期化後に呼び出す
func (p *Piece) set(x int, y int) {
	p.pos = newPos(x, y)
}

// 駒の文字列を作成
func (p *Piece) createSource() string {

	m := p.typ.Mark()
	b := m[0]
	if p.turn == TurnWhite {
		//小文字化
		b = b + 32
	}
	g := ""
	if p.growth {
		g = "+"
	}
	return string(b) + g
}

// 駒を示す文字列を取得
func (p *Piece) Mark() string {
	if p == nil {
		return ""
	}
	return p.source
}

// 駒がない状態
func (p *Piece) Empty() bool {
	if p == nil {
		return true
	}
	return false
}

type PieceType int

const (
	Pawn   PieceType = 0
	Lance  PieceType = 1
	Knight PieceType = 2
	Silver PieceType = 3
	Gold   PieceType = 4
	Rook   PieceType = 5
	Bishop PieceType = 6
	King   PieceType = 7

	GrowthPawn   PieceType = 8
	GrowthLance  PieceType = 9
	GrowthKnight PieceType = 10
	GrowthSilver PieceType = 11
	GrowthRook   PieceType = 12
	GrowthBishop PieceType = 13

	PieceTypeNotFound PieceType = 99
)

func (t PieceType) Value() int {
	switch t {
	case Pawn:
		return 100
	case GrowthPawn:
		return 600
	case Lance:
		return 300
	case GrowthLance:
		return 600
	case Knight:
		return 400
	case GrowthKnight:
		return 600
	case Silver:
		return 500
	case GrowthSilver:
		return 600
	case Gold:
		return 600
	case Rook:
		return 1000
	case GrowthRook:
		return 1500
	case Bishop:
		return 800
	case GrowthBishop:
		return 1000
	case King:
		return (Pawn.Value()*9+Lance.Value()*2+Knight.Value()*2+
			Silver.Value()*2+Gold.Value()*2+Rook.Value()+Bishop.Value())*2 + 1
	}
	return 0
}

// 駒文字(大文字=先手/小文字=後手)から駒種と手番を得る。
// 文字マッピングの仕様は core/sfen に集約している。
func NewPieceType(v byte) (PieceType, TurnType) {
	base, black := sfen.ParsePieceLetter(v)
	t := TurnBlack
	if !black {
		t = TurnWhite
	}
	if base == sfen.NotFound {
		return PieceTypeNotFound, t
	}
	return PieceType(base), t
}

// 駒種に対応する SFEN の大文字(成駒はベース駒の文字)を返す。
// 文字マッピングの仕様は core/sfen に集約している。
func (t PieceType) Mark() string {
	return sfen.Letter(int(t))
}

func (t PieceType) Base() PieceType {
	switch t {
	case GrowthPawn:
		return Pawn
	case GrowthLance:
		return Lance
	case GrowthKnight:
		return Knight
	case GrowthSilver:
		return Silver
	case GrowthRook:
		return Rook
	case GrowthBishop:
		return Bishop
	}
	return t
}

func (t PieceType) Growth() PieceType {
	switch t {
	case Pawn:
		return GrowthPawn
	case Lance:
		return GrowthLance
	case Knight:
		return GrowthKnight
	case Silver:
		return GrowthSilver
	case Rook:
		return GrowthRook
	case Bishop:
		return GrowthBishop
	}
	return PieceTypeNotFound
}

func (t PieceType) IsGrowth() bool {
	if t >= 8 {
		return true
	}
	return false
}

// 持っている駒を表現
// 解析用に王も入れておく
type Pieces [8]int

func NewPieces() Pieces {
	var v [8]int
	return v
}

func (p Pieces) copy() Pieces {
	return p
}

func (p *Pieces) add(t PieceType) {
	b := t.Base()
	p[b]++
}

func (p *Pieces) remove(t PieceType) {
	b := t.Base()
	p[b]--
}

func (p *Pieces) all() []PieceType {
	var rtn []PieceType
	//持っている場合追加
	for idx, v := range p {
		if v > 0 {
			rtn = append(rtn, PieceType(idx))
		}
	}
	return rtn
}

func (p Pieces) String() string {
	var b strings.Builder
	for idx := 0; idx < len(p); idx++ {
		if p[idx] != 0 {
			b.WriteString(fmt.Sprintf("%s", strings.Repeat(PieceType(idx).Mark(), p[idx])))
		}
	}
	return b.String()
}
