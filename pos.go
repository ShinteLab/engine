package shogi

import "github.com/ShinteLab/core/usi"

// 盤面位置情報
type Pos [2]uint8

var PosNone = newPos(0, 0)

func newPos(x, y int) Pos {
	var p Pos
	p[0] = uint8(x)
	p[1] = uint8(y)
	return p
}

// USI マス文字列を Pos に変換する。パース不能なら PosNone。
// 座標変換の仕様は core/usi に集約している。
func parsePos(buf string) Pos {
	x, y, ok := usi.ParseSquare(buf)
	if !ok {
		return PosNone
	}
	return newPos(x, y)
}

func (p Pos) XY() (int, int) {
	return int(p[0]), int(p[1])
}

func (p Pos) String() string {
	return usi.FormatSquare(int(p[0]), int(p[1]))
}
