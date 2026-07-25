package shogi_test

import (
	"fmt"
	"math/rand"
	"github.com/ShinteLab/engine"
	"testing"
)

// 同一 SFEN から作った2つの Board の Hash が一致することを確認する。
func TestHashSameSFENMatches(t *testing.T) {
	b1, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	b2, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if b1.Hash() != b2.Hash() {
		t.Errorf("hash mismatch for identical SFEN: %d != %d", b1.Hash(), b2.Hash())
	}
}

// 手順違いで同一局面に合流したとき Hash が一致することを確認する
// (7g7f 3c3d 2g2f と 2g2f 3c3d 7g7f)。
func TestHashTranspositionMatches(t *testing.T) {
	b1, err := shogi.NewBoard("startpos moves 7g7f 3c3d 2g2f")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	b2, err := shogi.NewBoard("startpos moves 2g2f 3c3d 7g7f")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if b1.Hash() != b2.Hash() {
		t.Errorf("hash mismatch for transposed move order: %d != %d", b1.Hash(), b2.Hash())
	}
}

// 異なる局面で Hash が不一致になることを確認する。
func TestHashDifferentPositionsMismatch(t *testing.T) {
	b1, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	b2, err := shogi.NewBoard("startpos moves 7g7f")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if b1.Hash() == b2.Hash() {
		t.Errorf("expected different hashes for different positions, both were %d", b1.Hash())
	}
}

// ランダムに合法手を進めながら、Copy() したハッシュと元の Board を
// 同じ手順で進めたハッシュが一致し続けることを確認する。
// (差分更新ではなく毎回全再計算なので、Copy 直後・Action 後どちらでも
// 常に正しいハッシュが得られるはずというテスト)
func TestHashCopyMatchesReplay(t *testing.T) {
	b, err := shogi.NewBoard(shogi.StartPos)
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	for i := 0; i < 20; i++ {
		actions := b.Candidate()
		if len(actions) == 0 {
			break
		}
		a := actions[i%len(actions)]

		copied := b.Copy()
		if copied.Hash() != b.Hash() {
			t.Fatalf("step %d: copy hash %d != original hash %d", i, copied.Hash(), b.Hash())
		}

		if !copied.Action(a) {
			t.Fatalf("step %d: action %s failed to apply on copy", i, a.String())
		}
		if !b.Action(a) {
			t.Fatalf("step %d: action %s failed to apply on original", i, a.String())
		}

		if copied.Hash() != b.Hash() {
			t.Fatalf("step %d: after applying %s, copy hash %d != original hash %d", i, a.String(), copied.Hash(), b.Hash())
		}
	}
}

// Stage I: DoMove の Zobrist 差分更新が computeHash() の全再計算と
// 常に一致することを、ランダムプレイアウト(固定シード、50手×10ゲーム)で
// 検証する。さらに UndoMove 後に盤面文字列・hash・turn・num が
// DoMove 前と完全一致することも確認する。
func TestDoMoveUndoMoveHashMatchesRecompute(t *testing.T) {

	r := rand.New(rand.NewSource(20240719))

	const numGames = 10
	const maxPlies = 50

	for g := 0; g < numGames; g++ {

		b, err := shogi.NewBoard(shogi.StartPos)
		if err != nil {
			t.Fatalf("NewBoard() error: %v", err)
		}

		for p := 0; p < maxPlies; p++ {

			actions := b.Candidate()
			if len(actions) == 0 {
				break
			}
			a := actions[r.Intn(len(actions))]

			//DoMove前のスナップショット
			prevBoardStr := fmt.Sprintf("%#v", b)
			prevHash := b.Hash()
			prevTurn := b.Turn()

			u, ok := b.DoMove(a)
			if !ok {
				t.Fatalf("game %d ply %d: DoMove(%s) failed", g, p, a.String())
			}

			//差分更新されたHashが全再計算と一致するか
			want := shogi.ExportBoardComputeHash(b)
			if b.Hash() != want {
				t.Fatalf("game %d ply %d: after DoMove(%s) hash=%d, want %d (recomputed)",
					g, p, a.String(), b.Hash(), want)
			}

			//UndoMoveで完全に元へ戻るか
			b.UndoMove(u)

			if fmt.Sprintf("%#v", b) != prevBoardStr {
				t.Fatalf("game %d ply %d: UndoMove(%s) did not restore board state\nbefore:\n%s\nafter:\n%s",
					g, p, a.String(), prevBoardStr, fmt.Sprintf("%#v", b))
			}
			if b.Hash() != prevHash {
				t.Fatalf("game %d ply %d: UndoMove(%s) hash=%d, want %d", g, p, a.String(), b.Hash(), prevHash)
			}
			if b.Turn() != prevTurn {
				t.Fatalf("game %d ply %d: UndoMove(%s) turn=%v, want %v", g, p, a.String(), b.Turn(), prevTurn)
			}

			//実際に手を進めて次のplyへ(DoMove/UndoMoveの繰り返しによる
			//スライス容量の使い回しが壊れないことも合わせて確認する)
			u2, ok := b.DoMove(a)
			if !ok {
				t.Fatalf("game %d ply %d: re-DoMove(%s) failed", g, p, a.String())
			}
			_ = u2
		}
	}
}
