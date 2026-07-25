package samples

import (
	"math/rand"
	"shinte/engine"
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

func (e *MoveEngine) GetVersion() string {
	return "0.0.0"
}

func (e *MoveEngine) GetAuthor() string {
	return "secondarykey"
}

func (e *MoveEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {

	pos := b.Candidate()
	if len(pos) == 0 {
		return shogi.ResignAction(), nil
	}

	n := rand.Intn(len(pos))
	a := pos[n]
	return a, nil

	//return shogi.Resign, nil
}
