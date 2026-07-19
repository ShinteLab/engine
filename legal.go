package shogi

// 合法手判定の軽量化(ピン検出方式)。
//
// 従来の legalCandidateSlow は擬似合法手1手ごとに copyLite+actionLite+IsCheck
// (敵全駒の利き再計算)を行っていた。ここでは局面につき1回だけ
// 「敵利きマップ・王手駒・ピン駒とその可動ライン」を事前計算し、
// 各擬似合法手を O(1)〜軽量なビット判定だけでふるいにかける。

// 合法手判定用の事前計算結果(手番側の玉について)
type legalInfo struct {
	hasKing bool
	kingSq  int

	//自玉を占有から除いて計算した敵利き集合(玉の移動先判定用)
	dangerMap BitBoard

	//自玉に王手している敵駒の集合
	checkers     BitBoard
	checkerCount int
	//王手駒が1枚のときの捕獲・合駒可能マス(捕獲先を含む)
	blockSquares BitBoard

	//ピンされている自駒の集合
	pinned BitBoard
	//ピンされている自駒ごとの可動マス集合(玉〜ピン駒の間+ピン駒の捕獲マス)
	pinLines map[int]BitBoard
}

// 現局面(手番側)の legalInfo を計算する。
func (b *Board) computeLegalInfo() legalInfo {

	own := b.camps[b.turn.Index()]
	enemy := own.enemy

	var info legalInfo

	kingBit := own.typBoards[King]
	if kingBit.isZero() {
		return info
	}
	info.hasKing = true

	kingSq := -1
	kingBit.forEach(func(sq int) {
		kingSq = sq
	})
	info.kingSq = kingSq

	//敵利きマップ: 自玉を占有から除いた占有で計算する
	//(走り駒の利きが玉を貫通する必要があるため。玉が利き筋に沿って
	// 後退する手を違法にする)
	occWithKing := own.board
	occWithKing.or(&enemy.board)

	occNoKing := occWithKing
	occNoKing.andNot(&kingBit)

	info.dangerMap = enemy.attackAll(&occNoKing)

	//王手駒の検出(利き逆算): 玉位置から各駒種の利きを求め、
	//敵の該当typBoardsとANDを取る。
	for t := 0; t < len(enemy.typBoards); t++ {
		bit := &enemy.typBoards[t]
		if bit.isZero() {
			continue
		}
		a := attacks(own.ownTurn, PieceType(t), kingSq, &occWithKing)
		a.and(bit)
		if !a.isZero() {
			info.checkers.or(&a)
		}
	}

	count := 0
	checkerSq := -1
	info.checkers.forEach(func(sq int) {
		count++
		checkerSq = sq
	})
	info.checkerCount = count

	if count == 1 {
		x, y := sqXY(checkerSq)
		p := enemy.Get(x, y)
		t := PieceTypeNotFound
		if p != nil {
			t = p.Type()
		}
		info.blockSquares = checkerBlockSquares(kingSq, checkerSq, t)
	}

	info.pinned, info.pinLines = computePins(kingSq, own, enemy, enemy.ownTurn)

	return info
}

// 王手駒(checkerSq, 種別t)に対する捕獲・合駒可能マスを求める。
// 捕獲先(checkerSq自身)は常に含む。走り駒(香・飛・角・竜・馬)であれば、
// 玉〜王手駒の間のマスも含める。桂馬など走らない駒による王手や、
// 竜・馬の隣接(王型)部分による王手は間駒不可のため捕獲先のみとなる
// (findDirectionが玉から王手駒への直線方向を見つけられない、または
//  隣接のため「間」のマスが存在しないケースとして自然に扱われる)。
func checkerBlockSquares(kingSq, checkerSq int, t PieceType) BitBoard {

	var result BitBoard
	result.setSq(checkerSq)

	switch t {
	case Lance, Rook, Bishop, GrowthRook, GrowthBishop:
		dir := findDirection(kingSq, checkerSq)
		if dir >= 0 {
			for _, s8 := range rayTable[kingSq][dir] {
				sq := int(s8)
				if sq == checkerSq {
					break
				}
				result.setSq(sq)
			}
		}
	}

	return result
}

// kingSq から targetSq への8方向レイのうち、targetSq を含む方向を返す。
// 見つからなければ -1(桂馬の利き等、直線上にない場合)。
func findDirection(kingSq, targetSq int) int {
	for dir := 0; dir < 8; dir++ {
		for _, s8 := range rayTable[kingSq][dir] {
			if int(s8) == targetSq {
				return dir
			}
		}
	}
	return -1
}

// 方向dirの走り駒として種別tが該当するか(ピン・間駒判定の対の判定に使う)。
// 香は手番(enemyTurn)ごとに攻撃方向が固定なので、対応する1方向のみ該当する。
func matchesSlider(t PieceType, dir int, enemyTurn TurnType) bool {
	switch t {
	case Rook, GrowthRook:
		return dir == dirN || dir == dirE || dir == dirS || dir == dirW
	case Bishop, GrowthBishop:
		return dir == dirNE || dir == dirSE || dir == dirSW || dir == dirNW
	case Lance:
		if enemyTurn == TurnBlack {
			return dir == dirS
		}
		return dir == dirN
	}
	return false
}

// 玉から8方向にレイを飛ばし、「自駒がちょうど1枚だけ」あってその先に
// 対応する敵の走り駒がある場合、その自駒をピンとして検出する。
func computePins(kingSq int, own, enemy *CampBoard, enemyTurn TurnType) (BitBoard, map[int]BitBoard) {

	var pinned BitBoard
	lines := make(map[int]BitBoard)

	for dir := 0; dir < 8; dir++ {

		ray := rayTable[kingSq][dir]
		blockerSq := -1

		for i, s8 := range ray {
			sq := int(s8)

			if own.board.getSq(sq) {
				if blockerSq == -1 {
					blockerSq = sq
					continue
				}
				//自駒が2枚目: このラインにピンは無い
				break
			}

			if enemy.board.getSq(sq) {
				if blockerSq != -1 {
					x, y := sqXY(sq)
					p := enemy.Get(x, y)
					if p != nil && matchesSlider(p.Type(), dir, enemyTurn) {
						pinned.setSq(blockerSq)

						var line BitBoard
						for j := 0; j <= i; j++ {
							line.setSq(int(ray[j]))
						}
						lines[blockerSq] = line
					}
				}
				//自駒(があれば)の先の最初の駒で方向の探索は終わり
				break
			}
			//空マスは読み進める
		}
	}

	return pinned, lines
}

// 擬似合法手 a が legalInfo に照らして合法かを判定する。
func (info *legalInfo) isLegal(a *Action) bool {

	if a.Hit() {
		if info.checkerCount >= 2 {
			return false
		}
		if info.checkerCount == 1 {
			ax, ay := a.after.XY()
			return info.blockSquares.getSq(squareOf(ax, ay))
		}
		return true
	}

	bx, by := a.before.XY()
	sourceSq := squareOf(bx, by)
	ax, ay := a.after.XY()
	destSq := squareOf(ax, ay)

	if sourceSq == info.kingSq {
		//玉の移動: 移動先が敵利きに入っていなければ合法。
		//dangerMapは自玉を占有から除いて計算済みなので、
		//守られている王手駒を玉で取る手も正しく違法になる。
		return !info.dangerMap.getSq(destSq)
	}

	if info.checkerCount >= 2 {
		//両王手は玉の移動でしか解消できない
		return false
	}

	if info.checkerCount == 1 && !info.blockSquares.getSq(destSq) {
		return false
	}

	if info.pinned.getSq(sourceSq) {
		line, ok := info.pinLines[sourceSq]
		if !ok || !line.getSq(destSq) {
			return false
		}
	}

	return true
}
