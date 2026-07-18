package shogi

import (
	"fmt"
	"math/bits"
	"strings"
)

// 81マス(9x9)を [2]uint64 で表現するビットボード。
// マス番号 sq = (y-1)*9 + (x-1) (0..80)。
// board[0] が sq 0..63、board[1] が sq 64..80(17ビット)を保持する。
type BitBoard struct {
	board [2]uint64
}

// board[1] で有効な17ビット分のマスク(sq 64..80 に対応)
const bitBoardHighMask = (uint64(1) << 17) - 1

func NewBitBoard() *BitBoard {
	var b BitBoard
	return &b
}

// (x, y) からマス番号を算出
func squareOf(x, y int) int {
	return (y-1)*9 + (x - 1)
}

// アドレス箇所
func (b *BitBoard) address(x, y int) (int, int) {
	sq := squareOf(x, y)
	return sq >> 6, sq & 63
}

// ビット設定
func (b *BitBoard) set(x, y int) {
	idx, bit := b.address(x, y)
	b.board[idx] |= uint64(1) << uint(bit)
}

// 駒を消す
func (b *BitBoard) clear(x, y int) bool {

	idx, bit := b.address(x, y)
	mask := uint64(1) << uint(bit)
	if (b.board[idx] & mask) > 0 {
		b.board[idx] &^= mask
		return true
	}

	return false
}

// 存在する場合、true
func (b *BitBoard) get(x, y int) bool {
	idx, bit := b.address(x, y)
	mask := uint64(1) << uint(bit)
	return (b.board[idx] & mask) > 0
}

// マス番号を直接指定してビット設定
func (b *BitBoard) setSq(sq int) {
	idx := sq >> 6
	bit := sq & 63
	b.board[idx] |= uint64(1) << uint(bit)
}

// マス番号を直接指定して存在判定
func (b *BitBoard) getSq(sq int) bool {
	idx := sq >> 6
	bit := sq & 63
	return (b.board[idx] & (uint64(1) << uint(bit))) > 0
}

// 盤面の加算(OR)
func (b *BitBoard) or(o *BitBoard) {
	b.board[0] |= o.board[0]
	b.board[1] |= o.board[1]
}

// 盤面の論理積(AND)
func (b *BitBoard) and(o *BitBoard) {
	b.board[0] &= o.board[0]
	b.board[1] &= o.board[1]
}

// 盤面の差分(AND NOT): o に立っているビットを自分から消す
func (b *BitBoard) andNot(o *BitBoard) {
	b.board[0] &^= o.board[0]
	b.board[1] &^= o.board[1]
}

// ビットが1つも立っていないか
func (b *BitBoard) isZero() bool {
	return b.board[0] == 0 && b.board[1] == 0
}

// x位置に存在するか？
func (b *BitBoard) hasX(x int) bool {
	for y := 1; y <= 9; y++ {
		if b.get(x, y) {
			return true
		}
	}
	return false
}

// セットされているマス番号を走査してfnを呼ぶ
func (b *BitBoard) forEach(fn func(sq int)) {
	for idx := 0; idx < len(b.board); idx++ {
		v := b.board[idx]
		for v != 0 {
			bit := bits.TrailingZeros64(v)
			fn(idx*64 + bit)
			//最下位ビットを落とす
			v &= v - 1
		}
	}
}

// フラグに応じた箇所
func (b *BitBoard) is(f bool) []Pos {
	pos := make([]Pos, 0, 64)

	src := b
	if !f {
		//反転した盤面を作って走査する
		var inv BitBoard
		inv.board[0] = ^b.board[0]
		inv.board[1] = b.board[1] ^ bitBoardHighMask
		src = &inv
	}

	src.forEach(func(sq int) {
		x := sq%9 + 1
		y := sq/9 + 1
		pos = append(pos, newPos(x, y))
	})
	return pos
}

func (b *BitBoard) exists() []Pos {
	return b.is(true)
}

// 空の位置を取得
func (b *BitBoard) empty() []Pos {
	return b.is(false)
}

// デバッグ用の文字列
func (b *BitBoard) GoString() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n%064b", b.board[0]))
	builder.WriteString(fmt.Sprintf("\n%017b", b.board[1]))
	return builder.String()
}
