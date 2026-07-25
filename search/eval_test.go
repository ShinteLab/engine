package search_test

import (
	"shinte/engine"
	"shinte/engine/search"
	"testing"
)

// (a) 初期局面は完全に左右対称(先手・後手のPSTがミラーで一致する)なので、
// eval() は 0 になるはず。
func TestEvalStartPositionIsSymmetric(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if got := search.ExportEval(b); got != 0 {
		t.Errorf("expected eval() == 0 at the symmetric start position, got %d", got)
	}
}

// (b) 歩を1つ前進させると、位置評価(PST)の分だけ「進めた側」に有利な
// 評価になるはず。eval() は手番側視点なので、黒が7g7fと指した直後は
// 白番であり、白から見て(黒の前進が良い手だった分)わずかにマイナスに
// なることを確認する。
func TestEvalRewardsPawnAdvance(t *testing.T) {
	b, err := shogi.NewBoard("startpos moves 7g7f")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	got := search.ExportEval(b)
	if got >= 0 {
		t.Errorf("expected eval() to be negative from the side to move (white) after black's pawn advance, got %d", got)
	}
	//Material差分は無い(歩を進めただけで捕獲は無い)ので、差はPSTのみに
	//由来する。歩=100基準・最大±50程度のスケールなので、Materialの
	//価値を超えない小さな値であることも確認する。
	if got <= -100 {
		t.Errorf("expected a small PST-only difference (|eval| < 100), got %d", got)
	}
}
