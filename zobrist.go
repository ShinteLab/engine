package shogi

import "math/rand"

// Zobrist ハッシュ用の乱数テーブル。
// [手番.Index()][駒種][マス番号] / [手番.Index()][基本駒種][持駒枚数 0..18]
var zobristPiece [2][14][81]uint64
var zobristHand [2][8][19]uint64
var zobristTurn uint64

// 固定シードで生成する(再現性のため)。
const zobristSeed = 20240719

func init() {
	r := rand.New(rand.NewSource(zobristSeed))

	for t := 0; t < 2; t++ {
		for pt := 0; pt < 14; pt++ {
			for sq := 0; sq < 81; sq++ {
				zobristPiece[t][pt][sq] = r.Uint64()
			}
		}
		for pt := 0; pt < 8; pt++ {
			for cnt := 0; cnt < 19; cnt++ {
				zobristHand[t][pt][cnt] = r.Uint64()
			}
		}
	}

	zobristTurn = r.Uint64()
}
