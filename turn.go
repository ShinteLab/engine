package shogi

type TurnType string

const (
	TurnBlack TurnType = "b"
	TurnWhite TurnType = "w"
)

// ターン切り替え
func (t TurnType) Next() TurnType {
	if t == TurnBlack {
		return TurnWhite
	}
	return TurnBlack
}

// 盤面のインデックス値
func (t TurnType) Index() int {
	if t == TurnBlack {
		return 0
	}
	return 1
}

// 盤面の成り位置
func (t TurnType) GrowthArea(y0, y1 int) bool {
	if t == TurnBlack {
		if y0 <= 3 || y1 <= 3 {
			return true
		}
	} else {
		if y0 >= 7 || y1 >= 7 {
			return true
		}
	}
	return false
}
