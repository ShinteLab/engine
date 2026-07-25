package shogi_test

import (
	"shinte/engine"
	"sort"
	"testing"
)

// Stage A(利きテーブル化)で手生成ロジックを書き換えても、
// 擬似合法手集合(文字列化・ソート済み)が変わらないことを機械的に担保する
// ゴールデンテスト。
//
// Stage B で Board.Candidate() は合法手化(王手放置・自殺手・打ち歩詰め除外)
// されたため、ゴールデンの期待値(擬似合法手前提)とは意味が変わった。
// このテストの目的は「手生成の幾何ロジックが壊れていないか」なので、
// 引き続き擬似合法手である pseudoCandidate() を export_test.go 経由で見る。
//
// go test . -run Golden で実行できる。

// Board の擬似合法手を Action.String() でソート済み []string 化する。
func goldenCandidateStrings(b *shogi.Board) []string {
	actions := shogi.ExportBoardPseudoCandidate(b)
	strs := make([]string, 0, len(actions))
	for _, a := range actions {
		strs = append(strs, a.String())
	}
	sort.Strings(strs)
	return strs
}

func assertGolden(t *testing.T, name string, b *shogi.Board, want []string) {
	t.Helper()
	got := goldenCandidateStrings(b)

	if len(got) != len(want) {
		t.Errorf("%s: length mismatch got=%d want=%d\ngot=%v\nwant=%v", name, len(got), len(want), got, want)
		return
	}
	for idx := range got {
		if got[idx] != want[idx] {
			t.Errorf("%s: mismatch at %d got=%q want=%q\ngot=%v\nwant=%v", name, idx, got[idx], want[idx], got, want)
			return
		}
	}
}

func TestGoldenCandidateStartPos(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	want := goldenStartPos
	assertGolden(t, "StartPos", b, want)
}

func TestGoldenCandidateMidGame(t *testing.T) {
	b, err := shogi.NewBoard("sfen 1n1k1g3/1srg5/l2p2b2/ppp1p1pG1/4P4/PP1P3P1/2PG2P1+l/L6R1/BNS1K1SN1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	b.SetStatus("b", "SN3Pl2p", "80")
	want := goldenMidGame
	assertGolden(t, "MidGame", b, want)
}

func TestGoldenCandidateGrowthPieces(t *testing.T) {
	// 竜(GrowthRook)・馬(GrowthBishop)が盤上にある局面
	b, err := shogi.NewBoard("sfen 5+P3/k3K4/2+B2+P3/9/5+L1P1/9/9/9/1+n5g1 w - 266")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	want := goldenGrowthPieces
	assertGolden(t, "GrowthPieces", b, want)
}

func TestGoldenCandidateHandPieces(t *testing.T) {
	// 持ち駒に歩・香・桂がある局面
	b, err := shogi.NewBoard("sfen 4k4/9/9/9/9/9/9/9/4K4 b RB2G2S2N2L9P 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	want := goldenHandPieces
	assertGolden(t, "HandPieces", b, want)
}

func TestGoldenCandidateComplex(t *testing.T) {
	// 竜馬・行き所のない駒判定が絡む複雑な局面(can_test.go TestCandidate2 由来)
	b, err := shogi.NewBoard("sfen 1+N7/+P8/4p1+P1k/7+B1/4K4/4PP3/7+pg/9/+l2g1+p1+l1 w - 264")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	want := goldenComplex
	assertGolden(t, "Complex", b, want)
}
