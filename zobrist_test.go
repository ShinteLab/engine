package shogi_test

import (
	"shogi"
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
