package samples

import (
	"math/rand"
	"shogi"
	"time"
)

func init() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
}

type MoveEngine struct {
	cnt int
}

func (e *MoveEngine) GetName() string {
	return "1st move Engine"
}

func (e *MoveEngine) GetAuthor() string {
	return "secondarykey"
}

func (e *MoveEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {

	pos := b.Can(true)
	if len(pos) == 0 {
		return shogi.NewAction(shogi.Resign), nil
	}

	n := rand.Intn(len(pos))
	a := pos[n]
	return a, nil

	//return shogi.Resign, nil
}
