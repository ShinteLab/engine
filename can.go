package shogi

type Vector []Pos

// 方向を持った位置を作成
func NewVector(pos ...Pos) Vector {
	var v Vector
	v = pos
	return v
}

// 盤面内判定
func isArea(v int) bool {
	if v < 1 || v > 9 {
		return false
	}
	return true
}

// 歩の動作
// attack.go の buildAttackTable() がテーブル構築のために呼び出す。
func canPawn(pos Pos, direction int) []Vector {

	x, y := pos.XY()
	my := y + (1 * direction)
	if !isArea(my) {
		return nil
	}

	mp := newPos(x, my)
	vecs := make([]Vector, 1)
	vecs[0] = NewVector(mp)
	return vecs
}

// 桂馬動作
// attack.go の buildAttackTable() がテーブル構築のために呼び出す。
func canKnight(pos Pos, direction int) []Vector {

	var vecs []Vector
	x, y := pos.XY()

	my := y + (2 * direction)
	if !isArea(my) {
		return nil
	}

	mx1 := x - 1
	if isArea(mx1) {
		vecs = append(vecs, NewVector(newPos(mx1, my)))
	}
	mx2 := x + 1
	if isArea(mx2) {
		vecs = append(vecs, NewVector(newPos(mx2, my)))
	}
	return vecs
}

// 銀動作
// attack.go の buildAttackTable() がテーブル構築のために呼び出す。
func canSilver(pos Pos, direction int) []Vector {

	vecs := make([]Vector, 0, 5)
	x, y := pos.XY()

	mpY := y + (1 * direction)
	mmY := y + (-1 * direction)

	//前
	if isArea(mpY) {
		vecs = append(vecs, NewVector(newPos(x, mpY)))
	}
	//左
	mx1 := x - 1
	if isArea(mx1) {
		if isArea(mpY) {
			vecs = append(vecs, NewVector(newPos(mx1, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVector(newPos(mx1, mmY)))
		}
	}
	//右
	mx2 := x + 1
	if isArea(mx2) {
		if isArea(mpY) {
			vecs = append(vecs, NewVector(newPos(mx2, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVector(newPos(mx2, mmY)))
		}
	}
	return vecs
}

// 金動作
// attack.go の buildAttackTable() がテーブル構築のために呼び出す。
func canGold(pos Pos, direction int) []Vector {

	vecs := make([]Vector, 0, 6)
	x, y := pos.XY()
	mpY := y + (1 * direction)
	mmY := y + (-1 * direction)

	//前
	if isArea(mpY) {
		vecs = append(vecs, NewVector(newPos(x, mpY)))
	}
	//後
	if isArea(mmY) {
		vecs = append(vecs, NewVector(newPos(x, mmY)))
	}

	//左
	mx1 := x - 1
	if isArea(mx1) {
		vecs = append(vecs, NewVector(newPos(mx1, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVector(newPos(mx1, mpY)))
		}
	}
	//右
	mx2 := x + 1
	if isArea(mx2) {
		vecs = append(vecs, NewVector(newPos(mx2, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVector(newPos(mx2, mpY)))
		}
	}
	return vecs
}

// 王動作
// attack.go の buildAttackTable() がテーブル構築のために呼び出す。
func canKing(pos Pos) []Vector {

	vecs := make([]Vector, 0, 8)
	x, y := pos.XY()
	mpY := y + 1
	mmY := y + -1

	//前
	if isArea(mpY) {
		vecs = append(vecs, NewVector(newPos(x, mpY)))
	}
	//後
	if isArea(mmY) {
		vecs = append(vecs, NewVector(newPos(x, mmY)))
	}

	//左
	mx1 := x - 1
	if isArea(mx1) {
		vecs = append(vecs, NewVector(newPos(mx1, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVector(newPos(mx1, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVector(newPos(mx1, mmY)))
		}
	}
	//右
	mx2 := x + 1
	if isArea(mx2) {
		vecs = append(vecs, NewVector(newPos(mx2, y)))
		if isArea(mpY) {
			vecs = append(vecs, NewVector(newPos(mx2, mpY)))
		}
		if isArea(mmY) {
			vecs = append(vecs, NewVector(newPos(mx2, mmY)))
		}
	}
	return vecs
}
