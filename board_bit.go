package shogi

type BitBoard struct {
	typ   PieceType
	board [3]uint32
}

func NewBitBoard(t PieceType) *BitBoard {
	var b BitBoard
	b.typ = t
	return &b
}

func (b *BitBoard) copy() *BitBoard {
	dst := NewBitBoard(b.typ)
	for idx, v := range b.board {
		dst.board[idx] = v
	}
	return dst
}

// ビット設定
func (b *BitBoard) set(x, y int) {
	//slog.Debug(fmt.Sprintf("set[%d,%d]", x, y))
	idx, bit := b.address(x, y)
	//slog.Debug(fmt.Sprintf("  before:%027b", b.board[idx]))
	b.board[idx] += 1 << bit
	//slog.Debug(fmt.Sprintf("  after :%027b", b.board[idx]))
}

// アドレス箇所
func (b *BitBoard) address(x, y int) (int, int) {
	// 上、中、下 0,1,2
	idx := (y - 1) / 3

	//行番号
	line := y % 3
	if line == 0 {
		line = 3
	}
	//X位置
	bit := ((line - 1) * 9) + x - 1
	return idx, bit
}

// 存在する場合、true
func (b *BitBoard) get(x, y int) bool {
	idx, bit := b.address(x, y)
	target := b.board[idx]
	mask := uint32(1 << bit)
	return (target & mask) > 0
}

// 駒を消す
func (b *BitBoard) clear(x, y int) bool {
	idx, bit := b.address(x, y)
	target := b.board[idx]
	mask := uint32(1 << bit)
	if (target & mask) > 0 {
		b.board[idx] -= mask
		return true
	}
	return false
}
