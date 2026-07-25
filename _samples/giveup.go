package samples

import (
	"github.com/ShinteLab/engine"
)

type GiveupEngine struct {
}

func (e *GiveupEngine) GetName() string {
	return "Give up Engine"
}

func (e *GiveupEngine) GetVersion() string {
	return "0.0.0"
}

func (e *GiveupEngine) GetAuthor() string {
	return "secondarykey"
}

func (e *GiveupEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {
	return shogi.ResignAction(), nil
}
