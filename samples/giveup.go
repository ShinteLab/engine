package samples

import (
	"shogi"
)

type GiveupEngine struct {
}

func (e *GiveupEngine) GetName() string {
	return "Give up Engine"
}

func (e *GiveupEngine) GetAuthor() string {
	return "secondarykey"
}

func (e *GiveupEngine) GetBest(b *shogi.Board) (string, error) {
	return shogi.Resign, nil
}
