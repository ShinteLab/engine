package shogi_test

import (
	"github.com/ShinteLab/engine"
	"testing"
)

// 合法手生成(Board.Candidate())の正しさを検証する perft。
// depth 0 で 1 を返し、各深さで Candidate() の手を DoMove/UndoMove して
// 単一の Board を掘り下げながら再帰する(Stage I: make/unmake化)。
func perft(b *shogi.Board, depth int) int {
	if depth == 0 {
		return 1
	}

	actions := b.Candidate()
	if depth == 1 {
		return len(actions)
	}

	sum := 0
	for _, a := range actions {
		u, ok := b.DoMove(a)
		if !ok {
			continue
		}
		sum += perft(b, depth-1)
		b.UndoMove(u)
	}
	return sum
}

// divide: depth-1 の perft を各初手ごとに出す(デバッグ用)。
func perftDivide(b *shogi.Board, depth int) map[string]int {
	result := make(map[string]int)
	actions := b.Candidate()
	for _, a := range actions {
		u, ok := b.DoMove(a)
		if !ok {
			continue
		}
		result[a.String()] = perft(b, depth-1)
		b.UndoMove(u)
	}
	return result
}

func TestPerftDepth1(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	got := perft(b, 1)
	want := 30
	if got != want {
		t.Errorf("perft(1) = %d, want %d", got, want)
	}
}

func TestPerftDepth2(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	got := perft(b, 2)
	want := 900
	if got != want {
		t.Errorf("perft(2) = %d, want %d", got, want)
	}
}

func TestPerftDepth3(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	got := perft(b, 3)
	want := 25470
	if got != want {
		t.Errorf("perft(3) = %d, want %d", got, want)
	}
}

func TestPerftDepth4(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping perft(4) in short mode")
	}
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	got := perft(b, 4)
	want := 719731
	if got != want {
		t.Errorf("perft(4) = %d, want %d", got, want)
	}
}

// BenchmarkPerft2: 合法手生成の総合速度計測(初期局面 depth2)
// StageH: 112828 ns/op
// StageI: perft自体をDoMove/UndoMove化。83844 ns/op
func BenchmarkPerft2(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		perft(board, 2)
	}
}

// BenchmarkPerft3: 合法手生成の総合速度計測(初期局面 depth3)。
// StageE(局面履歴・千日手判定の追加)によるコスト増を追跡するために追加。
// StageH: 4003449 ns/op
// StageI: perft自体をDoMove/UndoMove化。2699573 ns/op
func BenchmarkPerft3(b *testing.B) {
	board, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		b.Fatalf("NewBoard() error: %v", err)
	}
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		perft(board, 3)
	}
}
