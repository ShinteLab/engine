package shogi_test

import (
	"fmt"
	"log/slog"
	"github.com/ShinteLab/engine"
	"testing"
)

// StageB: Candidate() は合法手化(王手放置・自殺手・打ち歩詰め除外)された。
// このテストは駒取り(敵玉を含む)を伴う擬似合法手の生成数・内容を検証する
// 目的で書かれており、意図的に王手放置(敵玉を取れる状態)を作って確認して
// いるため、擬似合法手である pseudoCandidate() を使うよう維持する。
func TestCandidate(t *testing.T) {

	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Errorf("newBoard error:%v", err)
	}

	actions := shogi.ExportBoardPseudoCandidate(b)
	if len(actions) != 30 {
		t.Errorf("first action %d != %d(%v)", 30, len(actions), actions)
	}

	slog.Info(fmt.Sprintf("%v", actions))

	a := shogi.NewAction("8h7h")
	f := b.Action(a)
	if !f {
		t.Errorf("action error:%v", a)
	}

	b, err = shogi.NewBoard("sfen 5+P3/k3K4/2+B2+P3/9/5+L1P1/9/9/9/1+n5g1 w - 266")
	if err != nil {
		t.Errorf("new Board error:%v", err)
	}

	actions = shogi.ExportBoardPseudoCandidate(b)
	if len(actions) != 11 {
		t.Errorf("checkmate action %d != %d(%v)", 11, len(actions), actions)
	}

	nb := b.Copy()
	nb.Action(shogi.NewAction("9b9c"))

	actions = shogi.ExportBoardPseudoCandidate(nb)
	if len(actions) != 37 {
		t.Errorf("checkmate action %d != %d(%v)", 37, len(actions), actions)
	}

	for _, na := range actions {

		if na.String() == "7c9c" {
			ene := na.Enemy()
			if ene == nil {
				t.Errorf("not nil enemy")
			} else if ene.Type() != shogi.King {
				t.Errorf("not king")
			}
		}
	}

	/*
		actions = b.Can(false)
		if len(actions) != 30 {
			t.Errorf("first action %d != %d(%v)", 30, len(actions), actions)
		}

		a = shogi.NewAction("2b1b")
		f = b.Action(a)
		if !f {
			t.Errorf("action error:%v", a)
		}
	*/
}

// StageB: Candidate() が合法手化されたことに伴い、王手放置を意図的に許容する
// この検証(相手が次に敵玉を取れるかどうかで王手放置を検出する)は
// pseudoCandidate() ベースのまま維持する。
func TestCandidate2(t *testing.T) {

	b := parse("sfen 1+N7/+P8/4p1+P1k/7+B1/4K4/4PP3/7+pg/9/+l2g1+p1+l1 w - 264")
	fmt.Printf("%#v\n", b)

	actions := shogi.ExportBoardPseudoCandidate(b)
	if len(actions) != 25 {
		t.Errorf("want %v got %v", 25, len(actions))
	}

	//next := []string("1c1d", "1c1b", "1c2c")
	for _, a := range actions {
		nb := b.Copy()
		nb.Action(a)

		nactions := shogi.ExportBoardPseudoCandidate(nb)
		f := a.String()

		for _, na := range nactions {
			e := na.Enemy()
			if e != nil {
				if e.Type() == shogi.King {
					f = ""
					break
				}
			}
		}

		if f != "" && f != "1c1b" && f != "1c2d" {
			t.Errorf("Not Found King error:[%#v]\n%#v", a, nb)
			t.Errorf("actions %v", nactions)
		}
	}

	// Output:
	//  |  9  8  7  6  5  4  3  2  1  |
	// -------------------------------|
	// 1|     N+                      |
	// 2|  P+                         |
	// 3|              p     P+    k  |
	// 4|                       B+    |
	// 5|              K              |
	// 6|              P  P           |
	// 7|                       p+ g  |
	// 8|                             |
	// 9|  l+       g     p+    l+    |
	// -------------------------------|
	//  Black:PPPPPPNNSSRB
	// *White:PPPPPLLNSSGGR
	//

}

func TestCanBishop(t *testing.T) {

	//9a9b (p)
	//6d 7c 8b [p]
	//6f 7g+ 7g 8h+ 8h 9i+ 9i
	//4d 3c 2b 1a 4f 3g+ 3g 2h+ 2h 1i+ 1i]
	b := parse("sfen p8/9/9/9/4b4/9/9/9/8P w - 1")
	a := b.Candidate()
	if len(a) != (21 + 1) {
		t.Errorf("bithop move error:%v", a)
	}

	b = parse("sfen p8/9/9/9/4+b4/9/9/9/8P w - 1")
	a = b.Candidate()
	//6e 4e 5d 5f 6d 7c 8b [p] 6f 7g 8h 9i 4d 3c 2b 1a 4f 3g 2h 1i
	if len(a) != (19 + 1) {
		t.Errorf("growth bishop move error:%v", a)
	}
}

// 成り(必須)のテスト
func TestCanGrowth(t *testing.T) {
	b := parse("sfen 9/9/9/9/3n5/3n5/5l3/1p7/s7P w - 1")
	//P 8i+ [8i]
	//L 4h+ 4h 4i+ [4i]
	//7g+ [7g] 5g+ [5g] 7h+ 7h 5h+ 5h
	a := b.Candidate()
	if len(a) != (10) {
		t.Errorf("required growth error:%v", a)
	}
}

// StageA: canLogic()/canLance()/canBishop()/canRook() は利きテーブル化(attack.go)
// により実行時パスから除外され削除した。canPawn/canKnight/canSilver/canGold/canKing
// は attackTable 構築に引き続き使うため残し、ベンチマークは直接呼び出しに変更する。
// BenchmarkCanLance/CanBishop/CanRook は対象関数の削除に伴い削除した。

// BenchmarkCanPawn-20             387154971                3.134 ns/op
func BenchmarkCanPawn(b *testing.B) {
	pos := shogi.ExportNewPos(5, 5)
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCanPawn(pos, -1)
	}
}

// BenchmarkCanKnight-20           375329115                3.190 ns/op
func BenchmarkCanKnight(b *testing.B) {
	pos := shogi.ExportNewPos(5, 5)
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCanKnight(pos, -1)
	}
}

// BenchmarkCanSilver-20           37306587                31.67 ns/op
func BenchmarkCanSilver(b *testing.B) {
	pos := shogi.ExportNewPos(5, 5)
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCanSilver(pos, -1)
	}
}

// BenchmarkCanGold-20             18016340                67.92 ns/op
func BenchmarkCanGold(b *testing.B) {
	pos := shogi.ExportNewPos(5, 5)
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCanGold(pos, -1)
	}
}

// BenchmarkCanKing-20             10539480               110.1 ns/op
// BenchmarkCanKing-20             15738385                74.65 ns/op
func BenchmarkCanKing(b *testing.B) {
	pos := shogi.ExportNewPos(5, 5)
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCanKing(pos)
	}
}

// StageA: 利きテーブルの単一入口 attacks() のベンチマーク(飛車=走り駒の代表)
func BenchmarkAttacksRook(b *testing.B) {
	sq := shogi.ExportSquareOf(5, 5)
	var occ shogi.BitBoard
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportAttacks(shogi.TurnBlack, shogi.Rook, sq, &occ)
	}
}

// BenchmarkEmptyVector-20                 1000000000               0.1297 ns/op
func BenchmarkEmptyVector(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		shogi.NewVector()
	}
}

// BenchmarkSingleVector-20                1000000000               0.1191 ns/op
func BenchmarkSingleVector(b *testing.B) {
	p := shogi.ExportNewPos(1, 1)
	for idx := 0; idx < b.N; idx++ {
		shogi.NewVector(p)
	}
}

// BenchmarkSliceVector-20                 1000000000               0.1208 ns/op
func BenchmarkSliceVector(b *testing.B) {
	p1 := shogi.ExportNewPos(1, 1)
	p2 := shogi.ExportNewPos(1, 2)
	pos := []shogi.Pos{p1, p2}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		shogi.NewVector(pos...)
	}
}

func BenchmarkCanFilter(b *testing.B) {

}

func BenchmarkCanFilterEnemy(b *testing.B) {

}
