package shogi_test

import (
	"github.com/ShinteLab/engine"
	"testing"
)

// 先手: 敵陣(1-3段目)に玉+10枚(歩9+歩1)、持駒 飛2金8 でちょうど28点。
// 10(盤上歩)+10(飛2*5)+8(金8*1)=28 で成立するはず。
func TestCanDeclareWinBlackExactly28Points(t *testing.T) {
	b, err := shogi.NewBoard("sfen PPPPPPPPP/P3K4/9/9/9/9/9/9/8k b 2R8G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	if !b.CanDeclareWin() {
		t.Errorf("expected CanDeclareWin() to be true at exactly 28 points (black)")
	}
}

// 同じ盤面で持駒を金7枚に減らすと27点となり、先手(必要28点)には届かず不成立。
func TestCanDeclareWinBlack27PointsFails(t *testing.T) {
	b, err := shogi.NewBoard("sfen PPPPPPPPP/P3K4/9/9/9/9/9/9/8k b 2R7G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	if b.CanDeclareWin() {
		t.Errorf("expected CanDeclareWin() to be false at 27 points for black (needs 28)")
	}
}

// 点数は28点あっても敵陣内の駒(玉除く)が9枚しかなければ不成立(10枚必要)。
func TestCanDeclareWinFailsWithNinePiecesInZone(t *testing.T) {
	// 盤上の歩を9枚(9点)にし、持駒を飛2金9(19点)にして合計28点は満たすが、
	// 敵陣内の駒(玉除く)は9枚のみ。
	b, err := shogi.NewBoard("sfen PPPPPPPP1/P3K4/9/9/9/9/9/9/8k b 2R9G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	if b.CanDeclareWin() {
		t.Errorf("expected CanDeclareWin() to be false with only 9 pieces in the zone")
	}
}

// 28点・10枚の条件を満たしていても、王手が掛かっていれば不成立。
func TestCanDeclareWinFailsWhileInCheck(t *testing.T) {
	b, err := shogi.NewBoard("sfen PPPPPPPPP/P3K4/9/9/9/9/9/9/4r3k b 2R8G 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	if !b.IsCheck(shogi.TurnBlack) {
		t.Fatalf("setup error: expected black to be in check")
	}
	if b.CanDeclareWin() {
		t.Errorf("expected CanDeclareWin() to be false while in check")
	}
}

// 後手: 敵陣(7-9段目)に玉+10枚(歩9+歩1)、持駒 飛2金7 でちょうど27点。
// 10(盤上歩)+10(飛2*5)+7(金7*1)=27 で後手の必要点数27に到達し成立するはず。
func TestCanDeclareWinWhiteExactly27Points(t *testing.T) {
	b, err := shogi.NewBoard("sfen 8K/9/9/9/9/9/9/p3k4/ppppppppp w 2r7g 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	if !b.CanDeclareWin() {
		t.Errorf("expected CanDeclareWin() to be true at exactly 27 points (white)")
	}
}

// 同じ盤面で持駒を金6枚に減らすと26点となり、後手(必要27点)には届かず不成立。
func TestCanDeclareWinWhite26PointsFails(t *testing.T) {
	b, err := shogi.NewBoard("sfen 8K/9/9/9/9/9/9/p3k4/ppppppppp w 2r6g 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}
	if b.CanDeclareWin() {
		t.Errorf("expected CanDeclareWin() to be false at 26 points for white (needs 27)")
	}
}
