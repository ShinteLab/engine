package shogi

type Vec []Pos

// 方向を持った位置を作成
func NewVec(pos ...Pos) Vec {
	var v Vec
	v = pos
	return v
}

func isArea(v int) bool {
	if v >= 1 && v <= 9 {
		return true
	}
	return false
}

// 移動位置を取得する
func can(t PieceType, pos Pos, own, enemy *CampBoard) []*Action {

	d := 1
	if own.ownTurn == TurnBlack {
		d = -1
	}

	//駒に応じた方向を持ったポジションを取得
	var vecs []Vec
	switch t {
	case Pawn:
		vecs = canPawn(pos, d)
	case Lance:
		vecs = canLance(pos, d)
	case Knight:
		vecs = canKnight(pos, d)
	case Silver:
		vecs = canSilver(pos, d)
	case Gold, GrowthPawn, GrowthLance, GrowthKnight, GrowthSilver:
		vecs = canGold(pos, d)
	case King:
		vecs = canKing(pos)
	case Rook:
		vecs = canRook(pos, false)
	case GrowthRook:
		vecs = canRook(pos, true)
	case Bishop:
		vecs = canBishop(pos, false)
	case GrowthBishop:
		vecs = canBishop(pos, true)
	}

	//すでに成っている
	g := t.IsGrowth()
	//移動用の文字列を作成
	pbuf := pos.String()

	actions := make([]*Action, 0, 16)
	for _, vec := range vecs {
		//作成する
		var wk []*Action
		//その方向に駒がないかを判定
		for _, pos := range vec {

			x, y := pos.XY()
			//自陣の駒がある場合
			p := own.Get(x, y)
			if p != nil {
				//動作できない
				break
			}

			//新しい動作を作成
			np := newPos(x, y)
			npBuf := np.String()
			a := NewAction(pbuf + npBuf)

			wk = append(wk, a)
			//敵を検索
			e := enemy.Get(x, y)
			if e != nil {
				a.SetEnemy(e)
			}

			wk = append(wk, a)
			//まだ成ってないで成る場所の場合
			if !g && a.GrowthArea(own.ownTurn) {
				//成を追加
				ga := NewAction(pbuf + npBuf + "+")
				ga.SetEnemy(e)
				wk = append(wk, ga)
			}

			if e != nil {
				//相手駒がある場合終了
				break
			}
		}

		//可能な箇所を入れる
		if wk != nil {
			actions = append(actions, wk...)
		}
	}
	return actions
}

func canPawn(pos Pos, direction int) []Vec {

	x, y := pos.XY()
	mp := newPos(x, y+(1*direction))

	vecs := make([]Vec, 1)
	vecs[0] = NewVec(mp)
	return vecs
}

func canLance(pos Pos, direction int) []Vec {

	x, y := pos.XY()

	var mp []Pos
	for {
		y = y + (1 * direction)
		if !isArea(y) {
			break
		}
		p := newPos(x, y)
		mp = append(mp, p)
	}

	vecs := make([]Vec, 1)
	vecs[0] = NewVec(mp...)
	return vecs
}

func canKnight(pos Pos, direction int) []Vec {

	var vecs []Vec
	x, y := pos.XY()

	my := y + (2 * direction)

	mx1 := x - 1
	if isArea(mx1) {
		vecs = append(vecs, NewVec(newPos(mx1, my)))
	}
	mx2 := x + 1
	if isArea(mx2) {
		vecs = append(vecs, NewVec(newPos(mx2, my)))
	}
	return vecs
}

func canSilver(pos Pos, direction int) []Vec {

	var vecs []Vec
	x, y := pos.XY()
	mpY := y + (1 * direction)
	mmY := y + (1 * direction)

	//前
	if isArea(mpY) {
		vecs = append(vecs, NewVec(newPos(x, mpY)))
	}
	//左
	mx1 := x - 1
	if isArea(mx1) {
		if isArea(mpY) {
			vecs = append(vecs, NewVec(newPos(mx1, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVec(newPos(mx1, mmY)))
		}
	}
	//右
	mx2 := x + 1
	if isArea(mx2) {
		if isArea(mpY) {
			vecs = append(vecs, NewVec(newPos(mx2, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVec(newPos(mx2, mmY)))
		}
	}
	return vecs
}

func canGold(pos Pos, direction int) []Vec {

	var vecs []Vec
	x, y := pos.XY()
	mpY := y + (1 * direction)
	mmY := y + (1 * direction)

	//前
	if isArea(mpY) {
		vecs = append(vecs, NewVec(newPos(x, mpY)))
	}
	//後
	if isArea(mmY) {
		vecs = append(vecs, NewVec(newPos(x, mmY)))
	}

	//左
	mx1 := x - 1
	if isArea(mx1) {
		vecs = append(vecs, NewVec(newPos(mx1, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVec(newPos(mx1, mpY)))
		}
	}
	//右
	mx2 := x + 1
	if isArea(mx2) {
		vecs = append(vecs, NewVec(newPos(mx2, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVec(newPos(mx2, mpY)))
		}
	}
	return vecs
}

func canKing(pos Pos) []Vec {

	var vecs []Vec
	x, y := pos.XY()
	mpY := y + 1
	mmY := y + -1

	//前
	if isArea(mpY) {
		vecs = append(vecs, NewVec(newPos(x, mpY)))
	}
	//後
	if isArea(mmY) {
		vecs = append(vecs, NewVec(newPos(x, mmY)))
	}

	//左
	mx1 := x - 1
	if isArea(mx1) {
		vecs = append(vecs, NewVec(newPos(mx1, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVec(newPos(mx1, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVec(newPos(mx1, mmY)))
		}
	}
	//右
	mx2 := x + 1
	if isArea(mx2) {
		vecs = append(vecs, NewVec(newPos(mx2, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVec(newPos(mx2, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVec(newPos(mx2, mmY)))
		}
	}
	return vecs
}

func canBishop(pos Pos, g bool) []Vec {

	x, y := pos.XY()
	var vecs []Vec

	for kx := -1; kx > 2; kx = kx + 2 {
		for ky := -1; ky > 2; ky = ky + 2 {
			var mp []Pos

			mx := x
			my := y

			for {
				mx = mx + (1 * kx)
				my = my + (1 * ky)
				if !isArea(mx) || !isArea(my) {
					break
				}
				mp = append(mp, newPos(mx, my))
			}

			if mp != nil {
				vecs = append(vecs, NewVec(mp...))
			}

			//上下左右の追加
			if g {
				if isArea(x+kx) && isArea(y+ky) {
					vecs = append(vecs, NewVec(newPos(x+kx, y+ky)))
				}
			}
		}
	}

	//上下左右を追加
	return vecs
}

func canRook(pos Pos, g bool) []Vec {

	x, y := pos.XY()
	var vecs []Vec

	if g {
		for kx := -1; kx > 2; kx = kx + 2 {
			for ky := -1; ky > 2; ky = ky + 2 {
				//斜め部分の追加
				if isArea(x+kx) && isArea(y+ky) {
					vecs = append(vecs, NewVec(newPos(x+kx, y+ky)))
				}
			}
		}
	}

	//左右
	for kx := -1; kx > 2; kx = kx + 2 {
		var mp []Pos
		mx := x
		for {
			mx = mx + (1 * kx)
			if !isArea(mx) {
				break
			}
			mp = append(mp, newPos(mx, y))
		}
		if mp != nil {
			vecs = append(vecs, NewVec(mp...))
		}
	}

	//上下
	for ky := -1; ky > 2; ky = ky + 2 {
		var mp []Pos
		my := y
		for {
			my = my + (1 * ky)
			if !isArea(my) {
				break
			}
			mp = append(mp, newPos(x, my))
		}
		if mp != nil {
			vecs = append(vecs, NewVec(mp...))
		}
	}

	return vecs
}
