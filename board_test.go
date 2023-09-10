package shogi_test

import (
	"fmt"
	"log/slog"
	"os"

	"shogi"
	"testing"
)

func TestMain(m *testing.M) {
	defer shogi.SetFileLogger(slog.LevelDebug, "board_test.log").Close()
	code := m.Run()
	os.Exit(code)
}

func TestNewBoard(t *testing.T) {
	_, err := shogi.NewBoard(shogi.FirstPosSFEN)
	if err != nil {
		t.Errorf("shogi.NewBoard() error: %v", err)
	}
}

func ExampleBoard() {

	b, err := shogi.NewBoard(shogi.FirstPosSFEN)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("2g2f"))
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("2c2d"))
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("2f2e"))
	fmt.Printf("%#v\n", b)

	//取る
	b.Action(shogi.NewAction("2d2e"))
	fmt.Printf("%#v\n", b)
	//違う歩を動かす
	b.Action(shogi.NewAction("5g5f"))
	fmt.Printf("%#v\n", b)

	b.Action(shogi.NewAction("2e2f"))
	fmt.Printf("%#v\n", b)

	//飛車を回す
	b.Action(shogi.NewAction("8h5h"))
	fmt.Printf("%#v\n", b)

	//陣地に入る
	b.Action(shogi.NewAction("2f2g+"))
	fmt.Printf("%#v\n", b)

	// Output:
	// |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p  p  p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|                             |
	// 7|  P  P  P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p  p  p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|     P                       |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|     p                       |
	// 5|                             |
	// 6|     P                       |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|     p                       |
	// 5|     P                       |
	// 6|                             |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|     p                       |
	// 6|                             |
	// 7|  P     P  P  P  P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:P
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|     p                       |
	// 6|              P              |
	// 7|  P     P  P     P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:P
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|     p        P              |
	// 7|  P     P  P     P  P  P  P  |
	// 8|     B                 R     |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:P
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|     p        P              |
	// 7|  P     P  P     P  P  P  P  |
	// 8|     B        R              |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	//  Black:
	// *White:P
	//
	//  |  1  2  3  4  5  6  7  8  9  |
	// -------------------------------|
	// 1|  l  n  s  g  k  g  s  n  l  |
	// 2|     r                 b     |
	// 3|  p     p  p  p  p  p  p  p  |
	// 4|                             |
	// 5|                             |
	// 6|              P              |
	// 7|  P  p+ P  P     P  P  P  P  |
	// 8|     B        R              |
	// 9|  L  N  S  G  K  G  S  N  L  |
	// -------------------------------|
	// *Black:
	//  White:P
}
