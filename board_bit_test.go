package shogi_test

import (
	"github.com/ShinteLab/engine"
	"testing"
)

// BenchmarkNewBitBoard-20         1000000000               0.1178 ns/op
func BenchmarkNewBitBoard(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		shogi.NewBitBoard()
	}
}

// Stage3: BitBoard.copy() を削除し値コピーに統一したため、
// このベンチマークは対象メソッドがなくなり削除。
// (旧) BenchmarkNewBitBoardCopy-20     79745876                13.59 ns/op
// (旧) BenchmarkNewBitBoardCopy-20     100000000               11.89 ns/op

// BenchmarkNewBitBoardIsTrue-20            6707098               179.3 ns/op
// Stage0-2: is() bit-scan via math/bits.TrailingZeros32
// BenchmarkNewBitBoardIsTrue-20           45524642                27.51 ns/op
// Stage3: BitBoard.parent除去
// BenchmarkNewBitBoardIsTrue-20           45185316                25.30 ns/op
// StageA: BitBoard を [2]uint64 化、is() を forEach(TrailingZeros64)ベースに書き換え
// BenchmarkNewBitBoardIsTrue-20           51163980                24.44 ns/op
func BenchmarkNewBitBoardIsTrue(b *testing.B) {
	board := shogi.NewBitBoard()
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportBitBoardIs(board, true)
	}
}

// BenchmarkNewBitBoardIsFalse-20           2428950               501.1 ns/op
// Stage0-2: is() bit-scan via math/bits.TrailingZeros32
// BenchmarkNewBitBoardIsFalse-20           2747383               438.6 ns/op
// Stage3: BitBoard.parent除去
// BenchmarkNewBitBoardIsFalse-20           2792212               435.3 ns/op
// StageA: BitBoard を [2]uint64 化、is() を forEach(TrailingZeros64)ベースに書き換え
// BenchmarkNewBitBoardIsFalse-20           2740879               428.3 ns/op
func BenchmarkNewBitBoardIsFalse(b *testing.B) {
	board := shogi.NewBitBoard()
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportBitBoardIs(board, false)
	}
}

// BenchmarkNewBitBoardGet-20              506025284                2.342 ns/op
func BenchmarkNewBitBoardGet(b *testing.B) {
	board := shogi.NewBitBoard()
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportBitBoardGet(board, 5, 5)
	}
}

// is() のビット走査化(TrailingZeros32)が address() のビット配置と
// 整合していることを、全81マスの set/exists, clear/empty で往復確認する。
func TestBitBoardIsRoundTrip(t *testing.T) {

	for x := 1; x <= 9; x++ {
		for y := 1; y <= 9; y++ {

			board := shogi.NewBitBoard()
			shogi.ExportBitBoardSet(board, x, y)

			exists := shogi.ExportBitBoardIs(board, true)
			found := false
			for _, p := range exists {
				px, py := p.XY()
				if px == x && py == y {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("set(%d,%d) not found in exists(): %v", x, y, exists)
			}

			shogi.ExportBitBoardClear(board, x, y)

			empty := shogi.ExportBitBoardIs(board, false)
			found = false
			for _, p := range empty {
				px, py := p.XY()
				if px == x && py == y {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("clear(%d,%d) not found in empty(): %v", x, y, empty)
			}
		}
	}
}
