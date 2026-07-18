package shogi_test

import (
	"shogi"
	"testing"
)

// 歩を動かした後、元のマスが打ち候補(空きマス)として現れることを確認する。
// BitBoard.clear() の親盤面への伝播ロジックが反転していたため、
// 修正前は集約盤面(CampBoard.board)に元のマスのビットが残留し、
// emptyPos() がそのマスを空きとして扱わず、打ち候補から漏れていた。
func TestCandidateAfterPawnMoveHasHitAtVacatedSquare(t *testing.T) {

	b, err := shogi.NewBoard("sfen lnsgkgsnl/1r5b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL b L 1 moves 7g7f 3c3d")
	if err != nil {
		t.Fatalf("shogi.NewBoard() error: %v", err)
	}

	actions := b.Candidate()
	found := false
	for _, a := range actions {
		if a.String() == "L*7g" {
			found = true
			break
		}
	}

	if !found {
		strs := make([]string, 0, len(actions))
		for _, a := range actions {
			strs = append(strs, a.String())
		}
		t.Errorf("Candidate() does not contain \"L*7g\": %v", strs)
	}
}

// BenchmarkEmptyCanBit-20          7109636               169.0 ns/op
func BenchmarkEmptyCanBit(b *testing.B) {

	camp := shogi.NewCampBoard(shogi.TurnBlack)
	enemy := shogi.NewCampBoard(shogi.TurnWhite)
	shogi.ExportCampBoardSetEnemy(camp, enemy)
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCampBoardCanBit(camp, shogi.Pawn)
	}
}

// BenchmarkCanBit-20                141464              8292 ns/op
// bitboard change getPos
// bitboard parent
// BenchmarkCanBit-20                 97546             12269 ns/op
// Stage0-2: occupied() aggregate check, newMoveAction, BitBoard.is bit-scan
// BenchmarkCanBit-20                323508              3795 ns/op
// Stage3: CampBoard盤面を値型化
// BenchmarkCanBit-20                321999              3851 ns/op
// StageA: 利きテーブル化(attacks()+forEachでcanFilterのベクトル走査を置換)
// BenchmarkCanBit-20               1743066               679.8 ns/op
func BenchmarkCanBit(b *testing.B) {

	camp := shogi.NewCampBoard(shogi.TurnBlack)
	enemy := shogi.NewCampBoard(shogi.TurnWhite)
	shogi.ExportCampBoardSetEnemy(camp, enemy)
	for x := 1; x <= 9; x++ {
		for y := 1; y <= 9; y++ {
			shogi.ExportCampBoardSet(camp, x, y, shogi.Pawn)
		}
	}

	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		shogi.ExportCampBoardCanBit(camp, shogi.Pawn)
	}
}
