package shogi_test

import (
	"shinte/engine"
	"testing"
)

// 双方の玉が2マスを往復するだけの局面で、4回同一局面に達したら
// (王手が絡まないので)千日手(Draw)になることを確認する。
// 3回目の時点では None であることも確認する。
func TestRepetitionDrawAfterFourOccurrences(t *testing.T) {

	b, err := shogi.NewBoard("sfen k8/9/9/9/4K4/9/9/9/9 b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	//1サイクル(4手): 黒玉往復・白玉往復で元の局面に戻る
	cycle := []string{"5e5d", "9a9b", "5d5e", "9b9a"}

	applyAndCheck := func(mov string, wantStatus shogi.RepetitionStatus, label string) {
		t.Helper()
		a := shogi.NewAction(mov)
		if a == nil {
			t.Fatalf("%s: NewAction(%s) returned nil", label, mov)
		}
		if !b.Action(a) {
			t.Fatalf("%s: Action(%s) failed", label, mov)
		}
		if got := b.Repetition(); got != wantStatus {
			t.Errorf("%s: Repetition() = %v, want %v", label, got, wantStatus)
		}
	}

	//1周目・2周目はまだ千日手にならない(出現2回目・3回目)
	for cy := 0; cy < 2; cy++ {
		for _, mov := range cycle {
			applyAndCheck(mov, shogi.RepetitionNone, "cycle")
		}
	}

	//3周目: 最後の1手で4回目の出現となりDrawになるはず
	for i, mov := range cycle {
		want := shogi.RepetitionNone
		if i == len(cycle)-1 {
			want = shogi.RepetitionDraw
		}
		applyAndCheck(mov, want, "final cycle")
	}
}

// 連続王手の千日手: 黒の飛車が交互に王手を掛け続け、白玉がその都度
// 逃げるだけの手順を4回繰り返すと、王手を掛け続けた黒(=現局面の手番側)の
// 負け(PerpetualLose)になることを確認する。
func TestRepetitionPerpetualLose(t *testing.T) {

	b, err := shogi.NewBoard("sfen 5R3/9/9/9/4k4/9/9/9/8K b - 1")
	if err != nil {
		t.Fatalf("NewBoard() error: %v", err)
	}

	//1サイクル(4手): 飛車が5筋/6筋を切り替えて王手し続け、白玉が逃げ続ける
	cycle := []string{"4a5a", "5e4e", "5a4a", "4e5e"}

	for cy := 0; cy < 3; cy++ {
		for i, mov := range cycle {
			a := shogi.NewAction(mov)
			if a == nil {
				t.Fatalf("cycle %d: NewAction(%s) returned nil", cy, mov)
			}
			if !b.Action(a) {
				t.Fatalf("cycle %d: Action(%s) failed", cy, mov)
			}

			last := cy == 2 && i == len(cycle)-1
			got := b.Repetition()
			if last {
				if got != shogi.RepetitionPerpetualLose {
					t.Errorf("final move: Repetition() = %v, want RepetitionPerpetualLose", got)
				}
				if b.Turn() != shogi.TurnBlack {
					t.Errorf("expected black (the checking side) to be to move, got %v", b.Turn())
				}
			} else {
				if got != shogi.RepetitionNone {
					t.Errorf("cycle %d move %d (%s): Repetition() = %v, want RepetitionNone", cy, i, mov, got)
				}
			}
		}
	}
}
