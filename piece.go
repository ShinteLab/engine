package shogi

import (
	"fmt"
	"strings"
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

func NewPieceType(v byte) (PieceType, TurnType) {
	//65-90 A-Z
	//97-122 a-z
	t := TurnBlack
	if v >= 97 && v <= 122 {
		v = v - 32
		t = TurnWhite
	}
	return parsePieceType(v), t
}

func parsePieceType(p byte) PieceType {
	switch p {
	case 'P':
		return Pawn
	case 'L':
		return Lance
	case 'N':
		return Knight
	case 'S':
		return Silver
	case 'G':
		return Gold
	case 'R':
		return Rook
	case 'B':
		return Bishop
	case 'K':
		return King
	default:
		return PieceTypeNotFound
	}
}

func (t PieceType) Mark() string {
	switch t {
	case Pawn, GrowthPawn:
		return "P"
	case Lance, GrowthLance:
		return "L"
	case Knight, GrowthKnight:
		return "N"
	case Silver, GrowthSilver:
		return "S"
	case Gold:
		return "G"
	case Rook, GrowthRook:
		return "R"
	case Bishop, GrowthBishop:
		return "B"
	case King:
		return "K"
	}
	return "None"
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
