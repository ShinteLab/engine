package samples

import (
	"fmt"
	"log/slog"
	"math/rand"
	"github.com/ShinteLab/engine"
	"time"
)

func init() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
}

type CheckmateEngine struct {
	cnt int
}

func (e *CheckmateEngine) GetName() string {
	return "Checkmate Engine"
}

func (e *CheckmateEngine) GetVersion() string {
	return "0.0.0"
}

func (e *CheckmateEngine) GetAuthor() string {
	return "secondarykey"
}

func (e *CheckmateEngine) GetBest(b *shogi.Board) (*shogi.Action, error) {

	actions := b.Candidate()
	oks := filter(b, actions)
	if len(oks) == 0 {
		return shogi.ResignAction(), nil
	}

	n := rand.Intn(len(oks))
	a := oks[n]
	return a, nil
}

// 次の手の後に王手がある場合、阻止する手のみ返す
func filter(b *shogi.Board, actions []*shogi.Action) []*shogi.Action {

	var newActions []*shogi.Action

	slog.Debug(fmt.Sprintf("Can:[%d]", len(actions)))

	//全操作を設定
	for _, a := range actions {

		slog.Info(fmt.Sprintf("Action:[%v]", a))

		n := b.Copy()
		//動作させる
		if !n.Action(a) {
			slog.Error(fmt.Sprintf("action error:%v", a))
			continue
		}

		//次の手番の可能性のある操作をすべて取得
		n_actions := n.Candidate()

		check := false

		//全手数回繰り返す
		for _, na := range n_actions {
			enemy := na.Enemy()
			if enemy != nil {
				//取れてる
				if enemy.Type() == shogi.King {
					slog.Debug(fmt.Sprintf("Next King:[%v]", na))
					check = true
					break
				}
			}
		}

		if !check {
			newActions = append(newActions, a)
		}
	}

	slog.Debug(fmt.Sprintf("Filter Can:[%d]", len(newActions)))
	return newActions
}
