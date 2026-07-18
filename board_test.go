package shogi_test

import (
	"fmt"
	"log"

	"shogi"
	"testing"
)

func TestNewBoard(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Errorf("shogi.NewBoard() error: %v", err)
	}

	if b == nil {
		t.Errorf("shogi.NewBoard() is not nil")
	}
}

func Example_stop() {

	b, err := shogi.NewBoard("sfen 1n1k1g3/1srg5/l2p2b2/ppp1p1pG1/4P4/PP1P3P1/2PG2P1+l/L6R1/BNS1K1SN1")
	if err != nil {
		log.Fatalf("NewBoard() error: %+v", err)
	}
	b.SetStatus("b", "SN3Pl2p", "80")
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("2i1g"))
	fmt.Printf("%#v\n", b)

	// Output:
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|     n     k     g           |
	// 2|     s  r  g                 |
	// 3|  l        p        b        |
	// 4|  p  p  p     p     p  G     |
	// 5|              P              |
	// 6|  P  P     P           P     |
	// 7|        P  G        P     l+ |
	// 8|  L                    R     |
	// 9|  B  N  S     K     S  N     |
	// -------------------------------|
	// *Black:PPPNS
	//  White:PPL
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|     n     k     g           |
	// 2|     s  r  g                 |
	// 3|  l        p        b        |
	// 4|  p  p  p     p     p  G     |
	// 5|              P              |
	// 6|  P  P     P           P     |
	// 7|        P  G        P     N  |
	// 8|  L                    R     |
	// 9|  B  N  S     K     S        |
	// -------------------------------|
	//  Black:PPPLNS
	// *White:PPL
}

func ExampleBoard() {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		log.Fatalf("NewBoard() error: %+v", err)
	}

	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("8g8f"))
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("8c8d"))
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("8f8e"))
	fmt.Printf("%#v\n", b)

	//取る
	b.Action(shogi.NewAction("8d8e"))
	fmt.Printf("%#v\n", b)
	//違う歩を動かす
	b.Action(shogi.NewAction("5g5f"))
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("8e8f"))
	fmt.Printf("%#v\n", b)

	//飛車を回す
	b.Action(shogi.NewAction("2h5h"))
	fmt.Printf("%#v\n", b)

	//陣地に入る
	b.Action(shogi.NewAction("8f8g+"))
	fmt.Printf("%#v\n", b)

	// Output:
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p  p  p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|                             |
	// 7|  P  P  P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p  p  p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|     P                       |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|     p                       |
	// 5|                             |
	// 6|     P                       |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|     p                       |
	// 5|     P                       |
	// 6|                             |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|     p                       |
	// 6|                             |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:P
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|     p                       |
	// 6|              P              |
	// 7|  P     P  P     P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:P
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|     p        P              |
	// 7|  P     P  P     P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:P
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|     p        P              |
	// 7|  P     P  P     P  P  P  P  |
	// 8|     B        R              |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:P
	//
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|              P              |
	// 7|  P  p+ P  P     P  P  P  P  |
	// 8|     B        R              |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:P
}

func Example_stop2() {

	b, err := shogi.NewBoard("sfen 4k4/1r5b1/9/9/9/9/9/1B5R1/4K4 b 2G2S2N2L9P2g2s2n2l9p 1")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", b)
	//panic(shogi.Actions(b.Can()))

	// Output:
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|              k              |
	// 2|     r                 b     |
	// 3|                             |
	// 4|                             |
	// 5|                             |
	// 6|                             |
	// 7|                             |
	// 8|     B                 R     |
	// 9|              K              |
	// -------------------------------|
	// *Black:PPPPPPPPPLLNNSSGG
	//  White:PPPPPPPPPLLNNSSGG
}

// BenchmarkCopy-20                 2437864               485.2 ns/op
// Stage0-2: clear/set fix, Pieces.copy value-copy, etc.
// BenchmarkCopy-20                 2099673               586.8 ns/op
// Stage3: BitBoard.parent除去、CampBoard盤面を値型化(ゼロアロケーション化)
// BenchmarkCopy-20                12265398               110.8 ns/op
// StageA: BitBoard を [2]uint64 化(変更対象外だが計測値を記録)
// BenchmarkCopy-20                10706227               127.0 ns/op
func BenchmarkCopy(b *testing.B) {
	board := parse(shogi.StartPos)
	for idx := 0; idx < b.N; idx++ {
		board.Copy()
	}
}

// BenchmarkCandidate-20             103194             11470 ns/op
// Stage0-2: BitBoard.is bit-scan, occupied() aggregate check, newMoveAction
// BenchmarkCandidate-20             286834              4239 ns/op
// Stage3: CampBoard盤面を値型化
// BenchmarkCandidate-20             295641              4145 ns/op
// StageA: 利きテーブル化(BitBoard u64化 + attackTable/rayTable + canBit書き換え)
// BenchmarkCandidate-20             590817              1895 ns/op
// StageB: 合法手化(擬似合法手ごとに Copy+Action+IsCheck を実施するため増加)
// BenchmarkCandidate-20             100372             11897 ns/op
func BenchmarkCandidate(b *testing.B) {
	board := parse(shogi.StartPos)
	for idx := 0; idx < b.N; idx++ {
		board.Candidate()
	}
}

// BenchmarkBoardSet-20            178783014                6.623 ns/op
func BenchmarkBoardSet(b *testing.B) {

	board := parse(shogi.StartPos)
	p := shogi.NewPieceFromString("P")

	for idx := 0; idx < b.N; idx++ {
		shogi.ExportBoardSet(board, 1, 1, p)
	}
}
