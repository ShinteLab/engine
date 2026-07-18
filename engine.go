package shogi

import "context"

type Engine interface {
	GetName() string
	GetVersion() string
	GetAuthor() string
	GetBest(*Board) (*Action, error)
}

// 探索の途中経過情報(USI の info 行に対応する)
type Info struct {
	Depth   int
	ScoreCP int
	Nodes   int64
	PV      []*Action
	Mate    int
}

// ctx によるキャンセル・info 通知に対応した拡張エンジンインターフェース。
// 実装していれば usi.go の go ハンドラがこちらを優先して使う。
type ContextEngine interface {
	Engine
	GetBestContext(ctx context.Context, b *Board, info func(Info)) (*Action, error)
}

// setoption を受け取れる拡張インターフェース
type OptionEngine interface {
	SetOption(name, value string)
}

// 詰将棋探索(go mate)に対応する拡張インターフェース。
// 詰みが無い場合は (nil, nil) を返す。
type MateEngine interface {
	GetMate(ctx context.Context, b *Board) ([]*Action, error)
}
