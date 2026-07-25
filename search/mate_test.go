package search_test

import (
	"context"
	"github.com/ShinteLab/engine"
	"github.com/ShinteLab/engine/search"
	"testing"
	"time"
)

// 手順を順に適用し、最終局面が「王手されており、かつ合法手が0」
// (=詰み)であることを検証する共通ヘルパ。
func assertMatesOut(t *testing.T, b *shogi.Board, moves []*shogi.Action) {
	t.Helper()

	nb := b
	for _, m := range moves {
		nb = nb.Copy()
		if !nb.Action(m) {
			t.Fatalf("failed to apply move %s", m.String())
		}
	}

	if len(nb.Candidate()) != 0 {
		t.Errorf("expected 0 legal moves in the final position, got %v", nb.Candidate())
	}
	if !nb.InCheck() {
		t.Errorf("expected the final position to be in check (checkmate)")
	}
}

// 1手詰め(頭金)。moves長さ1で、適用すると相手の合法手0になること。
func TestMateOneMove(t *testing.T) {
	b, err := shogi.NewBoard("sfen 3pkp3/3s1s3/4S4/9/9/9/9/9/9 b G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	moves, found, err := search.Mate(context.Background(), b, 5)
	if err != nil {
		t.Fatalf("Mate() error: %v", err)
	}
	if !found {
		t.Fatalf("expected a mate to be found")
	}
	if len(moves) != 1 {
		t.Errorf("expected a 1-move mate, got %d moves: %v", len(moves), moves)
	}

	assertMatesOut(t, b, moves)
}

// 3手詰めの典型形。飛車の王手で玉を逃げ場一つに追い、金打ちで詰ます。
func TestMateThreeMoves(t *testing.T) {
	b, err := shogi.NewBoard("sfen 3pk1p2/3p1pp2/9/9/9/9/9/9/R7K b G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	moves, found, err := search.Mate(context.Background(), b, 5)
	if err != nil {
		t.Fatalf("Mate() error: %v", err)
	}
	if !found {
		t.Fatalf("expected a mate to be found")
	}
	if len(moves) != 3 {
		t.Errorf("expected a 3-move mate, got %d moves: %v", len(moves), moves)
	}

	assertMatesOut(t, b, moves)
}

// 詰まない局面では found=false であること(初期局面は明らかに詰まない)。
func TestMateNotFound(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	moves, found, err := search.Mate(context.Background(), b, 3)
	if err != nil {
		t.Fatalf("Mate() error: %v", err)
	}
	if found {
		t.Errorf("expected no mate from the start position, got moves: %v", moves)
	}
}

// 王手の掛け方を間違えると詰まない局面で、Mate() が正しく詰む手順を
// 選んで返すことを確認する(手順適用検証で兼ねる)。
// 持駒に金だけでなく銀も持たせ、銀打ちなど詰まない王手の選択肢が
// 存在する中でも正しい詰み手順(G*5b)を見つけられることを確認する。
func TestMateChoosesCorrectMateNotJustAnyCheck(t *testing.T) {
	b, err := shogi.NewBoard("sfen 3pkp3/3s1s3/4S4/9/9/9/9/9/9 b GS 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	moves, found, err := search.Mate(context.Background(), b, 5)
	if err != nil {
		t.Fatalf("Mate() error: %v", err)
	}
	if !found {
		t.Fatalf("expected a mate to be found")
	}

	assertMatesOut(t, b, moves)
}

// タイムアウト: 既に期限切れのctxを渡すと、複雑な局面でも
// panicせずにerrが返ること。
func TestMateTimeoutReturnsErrorWithoutPanic(t *testing.T) {
	b, err := shogi.NewBoard("sfen 1+N7/+P8/4p1+P1k/7+B1/4K4/4PP3/7+pg/9/+l2g1+p1+l1 w - 264")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	moves, found, err := search.Mate(ctx, b, 9)
	if err == nil {
		t.Errorf("expected an error for an already-expired context")
	}
	if found {
		t.Errorf("expected found=false when cancelled, got moves: %v", moves)
	}
}
