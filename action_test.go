package shogi_test

import (
	"shogi"
	"testing"
)

func TestAction(t *testing.T) {
	shogi.NewAction("1h1g")
	shogi.NewAction("1g#")
}

// BenchmarkNewAction-20           49494740                25.20 ns/op
func BenchmarkNewAction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		shogi.NewAction("1a1b")
	}
}
