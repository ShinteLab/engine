package shogi

// 動作
type Action struct {
	resign bool

	before Pos
	after  Pos
	hit    bool
	growth bool

	//動作予想時に相手がafterに居たらいれる
	enemy *Piece
}

// 動作文字列から動作を作成
func NewAction(buf string) *Action {

	var m Action

	m.before = PosNone
	m.after = PosNone
	m.growth = false
	m.hit = false

	if buf == Resign {
		m.resign = true
		return &m
	}

	leng := len(buf)

	if leng == 2 {
		//元位置のみ
		m.before = parsePos(buf)
	} else if leng == 3 {
		//打ちのみ
		m.after = parsePos(buf[0:2])
		m.hit = true
	} else if leng >= 4 {
		//ならない移動
		m.before = parsePos(buf[0:2])
		m.after = parsePos(buf[2:4])
		if leng == 5 {
			//なる移動
			m.growth = true
		}
	}
	return &m
}

// 動作時に取った駒を設定
func (a *Action) SetEnemy(p *Piece) {
	a.enemy = p
}

// 成っていい場所かを判定
func (m Action) GrowthArea(turn TurnType) bool {

	if m.hit {
		return false
	}

	_, y0 := m.before.XY()
	_, y1 := m.after.XY()

	if turn == TurnBlack {
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

// 動作文字列に変換
func (m Action) String() string {

	if m.resign {
		return Resign
	}

	if m.hit {
		return m.after.String() + "*"
	}
	g := ""
	if m.growth {
		g = "+"
	}
	return m.before.String() + m.after.String() + g
}
