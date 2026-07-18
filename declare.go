package shogi

// 入玉宣言勝ち(27点法, CSAルール準拠)の判定。
// 現在の手番側について判定する。
//
// 条件:
//  1. 玉が敵陣3段目以内(先手なら y<=3、後手なら y>=7)にいる。
//  2. 手番側に王手が掛かっていない。
//  3. 玉を除き、敵陣3段目以内にある手番側の駒が10枚以上。
//  4. 点数計算: 敵陣内の駒(玉除く)+持駒で、飛・角(成含む)=5点、
//     その他(玉除く)=1点。先手は28点以上、後手は27点以上で成立。
//
// 盤上の敵陣外の駒は点数に含まれない。
func (b *Board) CanDeclareWin() bool {

	t := b.turn
	camp := b.camps[t.Index()]

	king := &camp.typBoards[King]
	if king.isZero() {
		return false
	}

	//1. 玉が敵陣3段目以内にいるか
	kingInZone := false
	king.forEach(func(sq int) {
		_, y := sqXY(sq)
		if inEnemyZone(t, y) {
			kingInZone = true
		}
	})
	if !kingInZone {
		return false
	}

	//2. 王手が掛かっていないか
	if b.IsCheck(t) {
		return false
	}

	//3・4. 敵陣内の駒(玉除く)の枚数・点数を集計する
	count := 0
	score := 0
	for pt := 0; pt < len(camp.typBoards); pt++ {
		if PieceType(pt) == King {
			continue
		}
		bit := &camp.typBoards[pt]
		if bit.isZero() {
			continue
		}
		pt := pt
		bit.forEach(func(sq int) {
			_, y := sqXY(sq)
			if !inEnemyZone(t, y) {
				return
			}
			count++
			score += declarePoint(PieceType(pt))
		})
	}

	if count < 10 {
		return false
	}

	//持駒の点数を加算する
	for pt, cnt := range camp.has {
		if cnt <= 0 {
			continue
		}
		score += declarePoint(PieceType(pt)) * cnt
	}

	need := 27
	if t == TurnBlack {
		need = 28
	}

	return score >= need
}

// y が手番 t にとっての敵陣3段目以内かどうか
func inEnemyZone(t TurnType, y int) bool {
	if t == TurnBlack {
		return y <= 3
	}
	return y >= 7
}

// 入玉宣言点数(飛・角(成含む)=5点、その他=1点)
func declarePoint(t PieceType) int {
	switch t.Base() {
	case Rook, Bishop:
		return 5
	}
	return 1
}
