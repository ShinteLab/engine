package shogi

type Engine interface {
	GetName() string
	GetAuthor() string
	GetBest(*Board) (*Action, error)
}
