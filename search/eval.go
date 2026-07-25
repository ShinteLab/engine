package search

import "shinte/engine"

// 評価関数: 駒得(Material)差分 + 駒位置(PST)差分。
// どちらも「手番側から見た自分 - 相手」の形で返す。
func eval(b *shogi.Board) int {
	t := b.Turn()
	e := t.Next()

	material := b.Material(t) - b.Material(e)
	positional := pstScore(b, t) - pstScore(b, e)

	return material + positional
}

// 駒位置評価(PST: Piece-Square Table)。
//
// pst は先手(Black)視点で sq(0..80, (y-1)*9+(x-1)) ごとの値を保持する。
// 後手の駒は sq'=80-sq(盤面180度回転)で同じテーブルを参照する
// (search.SquareOf 相当の変換は shogi.Board.PieceSquares が返す sq と
//  同じ規約に合わせている)。
//
// スケールは歩=100を基準に、最大でも±50程度に収める
// (駒得評価を位置評価が覆さないようにするため)。
var pst [14][81]int16

func init() {
	buildPST()
}

// PieceType の並びは shogi.Pawn=0 ... shogi.GrowthBishop=13 に対応する。
func buildPST() {
	for sq := 0; sq < 81; sq++ {

		x, y := pstXY(sq)

		//advance: 0(自陣最後段, y=9) .. 8(敵陣最奥, y=1)。
		//先手は y が減る方向に進む(can.go の d=-1 と同じ)。
		advance := 9 - y
		center := centerness(x) // 0(端)..4(中央)

		//歩: 前進するほど加点(成りに近づくほど価値が上がる)
		pst[shogi.Pawn][sq] = int16(advance * 5) // 0..40

		//香・桂: 前進しすぎると行き所が減っていくため減点。
		//中段より手前は0(不利にならない)。
		lnPenalty := lanceKnightPenalty(advance)
		pst[shogi.Lance][sq] = lnPenalty
		pst[shogi.Knight][sq] = lnPenalty

		//銀・金(および歩香桂銀の成り=金と同格の動き): 中央・前方でやや加点
		goldLike := int16(advance*2 + center*2) // 0..~24
		pst[shogi.Silver][sq] = goldLike
		pst[shogi.Gold][sq] = goldLike
		pst[shogi.GrowthPawn][sq] = goldLike
		pst[shogi.GrowthLance][sq] = goldLike
		pst[shogi.GrowthKnight][sq] = goldLike
		pst[shogi.GrowthSilver][sq] = goldLike

		//角・飛: 中央志向でわずかに加点(利き数の代理)。成ると少し上乗せ。
		pst[shogi.Bishop][sq] = int16(center * 3)
		pst[shogi.Rook][sq] = int16(center * 3)
		pst[shogi.GrowthBishop][sq] = int16(center*3 + 10)
		pst[shogi.GrowthRook][sq] = int16(center*3 + 10)

		//玉: 自陣(y=7..9)側で端寄りなら加点(囲いの代理)。
		//敵陣側(y<=6)は中立(0)。入玉評価は将来課題。
		pst[shogi.King][sq] = kingSafety(x, y)
	}
}

// マス番号から (x, y) を復元する(shogi.squareOf の逆変換と同じ規約)。
func pstXY(sq int) (int, int) {
	return sq%9 + 1, sq/9 + 1
}

// 中央度: x=5(中央)で4、x=1,9(端)で0。
func centerness(x int) int {
	d := x - 5
	if d < 0 {
		d = -d
	}
	return 4 - d
}

func lanceKnightPenalty(advance int) int16 {
	if advance <= 4 {
		return 0
	}
	return int16(-(advance - 4) * 8) // advance=8で-32
}

func kingSafety(x, y int) int16 {
	if y <= 6 {
		//敵陣寄り〜中央は中立
		return 0
	}
	edge := 4 - centerness(x) // 0(中央)..4(端)
	return int16(edge * 5)    // 0..20
}

// 手番 t 側の駒すべての PST 値を合算する。
func pstScore(b *shogi.Board, t shogi.TurnType) int {
	sum := 0
	b.PieceSquares(t, func(pt shogi.PieceType, sq int) {
		s := sq
		if t == shogi.TurnWhite {
			s = 80 - sq
		}
		sum += int(pst[pt][s])
	})
	return sum
}
