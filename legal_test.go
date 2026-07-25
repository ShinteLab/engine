package shogi_test

import (
	"fmt"
	"github.com/ShinteLab/engine"
	"testing"
)

// 王手放置の手が Candidate() に含まれず、回避手のみが残ることを確認する。
// 黒玉(5e)、白飛車(5a)。飛車が縦のラインで王手をかけている。
func TestLegalCheckEvasionOnly(t *testing.T) {

	b, err := shogi.NewBoard("sfen 4r4/9/9/9/4K4/9/9/9/8k b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if !b.IsCheck(shogi.TurnBlack) {
		t.Fatalf("expected black to be in check")
	}

	actions := b.Candidate()
	fmt.Printf("CheckEvasionOnly actions: %v\n", actions)

	illegal := map[string]bool{
		"5e5d": true, // 玉が同じ筋(飛車の利き)に残る
		"5e5f": true, // 同上
	}

	if len(actions) == 0 {
		t.Fatalf("expected at least one evasion move")
	}

	for _, a := range actions {
		s := a.String()
		if illegal[s] {
			t.Errorf("evasion move %s still leaves king on the checking file", s)
		}

		nb := b.Copy()
		if !nb.Action(a) {
			t.Fatalf("action %s failed to apply", s)
		}
		if nb.IsCheck(shogi.TurnBlack) {
			t.Errorf("action %s leaves own king in check", s)
		}
	}
}

// ピン: 自玉と敵飛車の間にいる自駒(銀)が、ラインを外れる移動を候補に含まない。
func TestLegalPinnedPieceCannotLeaveLine(t *testing.T) {

	b, err := shogi.NewBoard("sfen 4r4/9/9/4S4/4K4/9/9/9/8k b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if b.IsCheck(shogi.TurnBlack) {
		t.Fatalf("king should not be in check while the silver blocks the file")
	}

	actions := b.Candidate()
	fmt.Printf("PinnedPiece actions: %v\n", actions)

	offLine := map[string]bool{
		"5d6c": true, // 銀が斜めへ(筋を外れる)
		"5d4c": true,
		"5d6e": true,
		"5d4e": true,
	}

	for _, a := range actions {
		s := a.String()
		if offLine[s] {
			t.Errorf("pinned silver should not be able to move off the file: %s", s)
		}
	}

	// 銀がラインに沿って前進する手(5d5c)自体は許可されているはず
	found := false
	for _, a := range actions {
		if a.String() == "5d5c" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected pinned silver to still be able to advance along the pin line (5d5c): %v", actions)
	}
}

// 自玉を相手の利きに晒す玉移動が Candidate() に含まれないことを確認する。
// 黒玉(5e)。白飛車(1d)が5段目ではなく4段目を横に利かせている。
func TestLegalKingCannotMoveIntoAttack(t *testing.T) {

	b, err := shogi.NewBoard("sfen 9/9/9/r8/4K4/9/9/9/8k b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	if b.IsCheck(shogi.TurnBlack) {
		t.Fatalf("king should not currently be in check")
	}

	actions := b.Candidate()
	fmt.Printf("KingCannotMoveIntoAttack actions: %v\n", actions)

	attacked := map[string]bool{
		"5e6d": true, // (4,4) 4段目は飛車の利き
		"5e5d": true, // (5,4)
		"5e4d": true, // (6,4)
	}

	for _, a := range actions {
		s := a.String()
		if attacked[s] {
			t.Errorf("king should not be able to move into an attacked square: %s", s)
		}
		nb := b.Copy()
		if !nb.Action(a) {
			t.Fatalf("action %s failed to apply", s)
		}
		if nb.IsCheck(shogi.TurnBlack) {
			t.Errorf("action %s leaves king in check: %s", s, s)
		}
	}

	if len(actions) != 5 {
		t.Errorf("expected 5 legal king moves, got %d: %v", len(actions), actions)
	}
}

// 打ち歩詰め: 逃げ場のない白玉に黒が歩を打って詰ますのは禁じ手。
func TestLegalUchifuzumeExcluded(t *testing.T) {

	b, err := shogi.NewBoard("sfen 3pkp3/3s1s3/4G4/9/9/9/9/9/9 b P 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	pseudo := shogi.ExportBoardPseudoCandidate(b)
	pseudoHas := false
	for _, a := range pseudo {
		if a.String() == "P*5b" {
			pseudoHas = true
		}
	}
	if !pseudoHas {
		t.Fatalf("P*5b should be pseudo-legal (sanity check): %v", pseudo)
	}

	actions := b.Candidate()
	fmt.Printf("UchifuzumeExcluded actions: %v\n", actions)
	for _, a := range actions {
		if a.String() == "P*5b" {
			t.Errorf("P*5b is uchifuzume (pawn-drop mate) and must be excluded")
		}
	}
}

// 打ち歩詰めではない歩打ち(王手はかかるが玉に逃げ場がある)は許可される。
func TestLegalPawnDropCheckNotMateAllowed(t *testing.T) {

	b, err := shogi.NewBoard("sfen 3pk4/3s1s3/4G4/9/9/9/9/9/9 b P 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	actions := b.Candidate()
	fmt.Printf("PawnDropCheckNotMateAllowed actions: %v\n", actions)

	found := false
	for _, a := range actions {
		if a.String() == "P*5b" {
			found = true
		}
	}
	if !found {
		t.Errorf("P*5b delivers check but is not mate (king can flee to 6a), so it should be allowed: %v", actions)
	}
}
