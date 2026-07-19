package shogi

import (
	"math/rand"
	"sort"
	"testing"
)

// Stage H: ピン検出方式の新 legalCandidate と、従来の copyLite ベースの
// legalCandidateSlow の結果集合が完全一致することを、ランダムプレイアウトで
// 検証する。王手局面・ピン局面は指し手を進めるうちに自然に多数出現する。
func TestLegalCandidateMatchesSlow(t *testing.T) {

	r := rand.New(rand.NewSource(20240719))

	const numGames = 15
	const minPlies = 20
	const maxPlies = 80

	total := 0
	mismatches := 0

	for g := 0; g < numGames; g++ {

		b, err := NewBoard(StartPos)
		if err != nil {
			t.Fatalf("NewBoard() error: %v", err)
		}

		plies := minPlies + r.Intn(maxPlies-minPlies+1)

		for p := 0; p < plies; p++ {

			fast := b.legalCandidate(true)
			slow := b.legalCandidateSlow(true)
			total++

			fastStrs := sortedActionStrings(fast)
			slowStrs := sortedActionStrings(slow)

			if !equalStrings(fastStrs, slowStrs) {
				mismatches++
				t.Errorf("game %d ply %d: mismatch\nfast(%d)=%v\nslow(%d)=%v",
					g, p, len(fastStrs), fastStrs, len(slowStrs), slowStrs)
				if mismatches > 5 {
					t.Fatalf("too many mismatches (%d), stopping early", mismatches)
				}
			}

			if len(fast) == 0 {
				//詰み/ステイルメイト相当。この対局はここで終了。
				break
			}

			a := fast[r.Intn(len(fast))]
			if !b.Action(a) {
				t.Fatalf("game %d ply %d: failed to apply chosen move %s", g, p, a.String())
			}
		}
	}

	t.Logf("checked %d positions across %d games, %d mismatches", total, numGames, mismatches)
}

func sortedActionStrings(actions []*Action) []string {
	s := make([]string, 0, len(actions))
	for _, a := range actions {
		s = append(s, a.String())
	}
	sort.Strings(s)
	return s
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
