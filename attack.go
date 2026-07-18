package shogi

// 利きテーブル基盤。
// init() で全マス×全駒種×手番の利き情報を事前計算しておき、
// 探索のホットパスではテーブル参照とレイ走査のみで済ませる。
//
// 座標系: 先手(TurnBlack)は -y 方向に進む(can.go の d = -1 と同じ)。
// マス番号 sq = (y-1)*9 + (x-1) は board_bit.go の squareOf() と同一。

// 8方向。y±x±で定義する。
const (
	dirN = iota
	dirNE
	dirE
	dirSE
	dirS
	dirSW
	dirW
	dirNW
)

var dirOffsets = [8][2]int{
	{0, -1}, // N
	{1, -1}, // NE
	{1, 0},  // E
	{1, 1},  // SE
	{0, 1},  // S
	{-1, 1}, // SW
	{-1, 0}, // W
	{-1, -1},// NW
}

// 非走り駒の利きテーブル。
// [TurnType.Index()][PieceType][sq]
// Lance / Rook / Bishop はレイで扱うためここでは使わない。
// GrowthRook / GrowthBishop は「王型の1歩部分」(竜=斜め4方向、馬=上下左右4方向)のみを保持する。
var attackTable [2][14][81]BitBoard

// 走り駒のレイ。[sq][方向] にマス番号を原点に近い順で保持する。
var rayTable [81][8][]uint8

func init() {
	buildAttackTable()
	buildRayTable()
}

// マス番号から (x, y) を復元する
func sqXY(sq int) (int, int) {
	return sq%9 + 1, sq/9 + 1
}

// 移動位置の一覧(Vector群)を BitBoard に変換する
func vecToBitBoard(vecs []Vector) BitBoard {
	var bb BitBoard
	for _, vec := range vecs {
		for _, p := range vec {
			x, y := p.XY()
			bb.setSq(squareOf(x, y))
		}
	}
	return bb
}

// 非走り駒テーブルの構築。can.go の幾何関数をそのまま利用する。
func buildAttackTable() {

	for ti := 0; ti < 2; ti++ {
		turn := TurnBlack
		if ti == 1 {
			turn = TurnWhite
		}
		d := 1
		if turn == TurnBlack {
			d = -1
		}

		for sq := 0; sq < 81; sq++ {
			x, y := sqXY(sq)
			pos := newPos(x, y)

			attackTable[ti][Pawn][sq] = vecToBitBoard(canPawn(pos, d))
			attackTable[ti][Knight][sq] = vecToBitBoard(canKnight(pos, d))
			attackTable[ti][Silver][sq] = vecToBitBoard(canSilver(pos, d))

			gold := vecToBitBoard(canGold(pos, d))
			attackTable[ti][Gold][sq] = gold
			attackTable[ti][GrowthPawn][sq] = gold
			attackTable[ti][GrowthLance][sq] = gold
			attackTable[ti][GrowthKnight][sq] = gold
			attackTable[ti][GrowthSilver][sq] = gold

			king := canKing(pos)
			attackTable[ti][King][sq] = vecToBitBoard(king)

			//竜=斜め4方向、馬=上下左右4方向の「王型の1歩部分」に分割
			var diag, orth BitBoard
			for _, vec := range king {
				for _, p := range vec {
					px, py := p.XY()
					dx := px - x
					dy := py - y
					if dx != 0 && dy != 0 {
						diag.setSq(squareOf(px, py))
					} else {
						orth.setSq(squareOf(px, py))
					}
				}
			}
			attackTable[ti][GrowthRook][sq] = diag
			attackTable[ti][GrowthBishop][sq] = orth
		}
	}
}

// レイテーブルの構築
func buildRayTable() {
	for sq := 0; sq < 81; sq++ {
		x, y := sqXY(sq)
		for d := 0; d < 8; d++ {
			dx, dy := dirOffsets[d][0], dirOffsets[d][1]
			mx, my := x, y
			var ray []uint8
			for {
				mx += dx
				my += dy
				if !isArea(mx) || !isArea(my) {
					break
				}
				ray = append(ray, uint8(squareOf(mx, my)))
			}
			rayTable[sq][d] = ray
		}
	}
}

// 単一方向のレイの利きを求める。occ に当たったマスを含めて打ち切る。
func rayAttack(sq int, dir int, occ *BitBoard) BitBoard {
	var result BitBoard
	for _, s := range rayTable[sq][dir] {
		result.setSq(int(s))
		if occ.getSq(int(s)) {
			break
		}
	}
	return result
}

// 複数方向のレイの利きを合成する
func rayAttackMulti(sq int, dirs []int, occ *BitBoard) BitBoard {
	var result BitBoard
	for _, d := range dirs {
		ra := rayAttack(sq, d, occ)
		result.or(&ra)
	}
	return result
}

var rookDirs = []int{dirN, dirE, dirS, dirW}
var bishopDirs = []int{dirNE, dirSE, dirSW, dirNW}

// 単一入口: turn/駒種/マスから利き(移動可能マスの集合)を求める。
// 自駒除外は呼び出し側で andNot すること(ここでは行わない)。
func attacks(turn TurnType, t PieceType, sq int, occ *BitBoard) BitBoard {

	ti := turn.Index()

	switch t {
	case Lance:
		dir := dirN
		if turn == TurnWhite {
			dir = dirS
		}
		return rayAttack(sq, dir, occ)
	case Rook:
		return rayAttackMulti(sq, rookDirs, occ)
	case Bishop:
		return rayAttackMulti(sq, bishopDirs, occ)
	case GrowthRook:
		r := rayAttackMulti(sq, rookDirs, occ)
		r.or(&attackTable[ti][GrowthRook][sq])
		return r
	case GrowthBishop:
		r := rayAttackMulti(sq, bishopDirs, occ)
		r.or(&attackTable[ti][GrowthBishop][sq])
		return r
	default:
		return attackTable[ti][t][sq]
	}
}

// そのマスからその駒が(盤面が空でも)動けるかどうか。
// 歩・香・桂の行き所のない駒判定に使う。
func hasMoveMask(turn TurnType, t PieceType, sq int) bool {

	ti := turn.Index()

	switch t {
	case Lance:
		dir := dirN
		if turn == TurnWhite {
			dir = dirS
		}
		return len(rayTable[sq][dir]) > 0
	case Rook, Bishop, GrowthRook, GrowthBishop:
		//盤内であれば必ずいずれかの方向に動ける
		return true
	default:
		return !attackTable[ti][t][sq].isZero()
	}
}
