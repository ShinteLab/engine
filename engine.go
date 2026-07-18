package shogi

type Engine interface {
	GetName() string
	GetVersion() string
	GetAuthor() string
	GetBest(*Board) (*Action, error)
}
