package search

import (
	"math/rand"
	"shinte/engine"
	"sync"
	"testing"
)

// Stage K-1: 手の圧縮表現(encodeMove/decodeMove)の往復が、
// 通常の移動・成りの移動・打ちのそれぞれで元のAction.String()と
// 一致することを確認する。
func TestEncodeDecodeMoveRoundTrip(t *testing.T) {

	cases := []string{
		"7g7f",
		"8h2b+",
		"P*5e",
		"G*1a",
		"1a1b",
		"9i9a+",
	}

	for _, s := range cases {
		a := shogi.NewAction(s)
		if a == nil {
			t.Fatalf("NewAction(%s) returned nil", s)
		}

		encoded := encodeMove(a)
		decoded := decodeMove(encoded)

		if decoded == nil {
			t.Errorf("decodeMove() returned nil for %s", s)
			continue
		}
		if decoded.String() != s {
			t.Errorf("round-trip mismatch: %s -> %d -> %s", s, encoded, decoded.String())
		}
	}
}

// Stage K-1: 複数goroutineが同一のtransTableへ並行してstore/probeを
// 繰り返しても、probeが返す(check^data==keyが成立した)エントリの内容が
// 常にkeyと整合していることを確認するストレステスト。
//
// -race が使えない環境でも成立するよう、xor-verifyの整合性チェック
// そのものをアサートする: scoreをkeyから決定的に導出できる値にしておき、
// probeがok=trueを返した場合、そのscoreが「そのkeyについて書き込まれる
// はずの値」と完全に一致することを確認する。もしdataとcheckが別の
// 書き込みに由来する「tear」を起こしていれば、check^data!=keyとなって
// 安全側(miss)に倒れるはずで、誤ったscoreがok=trueとともに返ることは
// 無いはずである。
func TestTransTableConcurrentStoreProbeConsistency(t *testing.T) {

	tt := newTransTable()

	const workers = 8
	const iterations = 20000

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(seed int64) {
			defer wg.Done()

			r := rand.New(rand.NewSource(seed))

			for i := 0; i < iterations; i++ {

				key := r.Uint64()
				//scoreはkeyから決定的に導出する(mate閾値を超えないよう
				//十分小さい範囲に収める)。
				wantScore := int(int32(key & 0xFFFF))
				depth := int(key % 16)
				flag := ttFlag(key % 3)

				tt.store(key, depth, 0, wantScore, flag, nil)

				gotScore, _, ok := tt.probe(key, 0, 0, -(1 << 20), 1<<20)
				if ok && gotScore != wantScore {
					t.Errorf("worker seed=%d iter=%d: corrupted entry for key=%d: got score=%d, want=%d",
						seed, i, key, gotScore, wantScore)
					return
				}
			}
		}(int64(w))
	}

	wg.Wait()
}

// 同一 key に対する store の直後、必ず probe で読み戻せる
// (自分の書き込みが他goroutineに一切邪魔されないシングルスレッド条件下での
// 基本的な健全性)ことも合わせて確認する。
func TestTransTableStoreThenProbeSingleThreaded(t *testing.T) {

	tt := newTransTable()

	key := uint64(0x1234567890ABCDEF)
	tt.store(key, 5, 0, 123, ttExact, nil)

	score, _, ok := tt.probe(key, 5, 0, -(1 << 20), 1<<20)
	if !ok {
		t.Fatalf("expected probe to hit immediately after store")
	}
	if score != 123 {
		t.Errorf("expected score=123, got %d", score)
	}
}
