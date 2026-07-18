package search

import "shogi"

// 置換表のスコア種別
type ttFlag uint8

const (
	ttExact ttFlag = iota
	ttLower
	ttUpper
)

type ttEntry struct {
	key   uint64
	depth int8
	flag  ttFlag
	score int32
	best  *shogi.Action
}

// 固定サイズ配列 + インデックスマスクによる置換表。
// map ではなくスライスを使うことで GC 負荷を避ける。
// 衝突時は key 照合で検出し、常時上書き(replace)方式とする。
const ttSizeBits = 20
const ttSize = 1 << ttSizeBits
const ttMask = uint64(ttSize - 1)

type transTable struct {
	entries []ttEntry
}

func newTransTable() *transTable {
	return &transTable{entries: make([]ttEntry, ttSize)}
}

func (tt *transTable) index(key uint64) uint64 {
	return key & ttMask
}

// mate スコア(詰みが絡むスコア)かどうかの判定しきい値。
// 通常の駒価値評価(最大でも数万程度)より十分大きく、
// 現実的な探索深さ(せいぜい数十手)による ply 補正の幅より
// 十分小さいマージンを取っている。
const mateThreshold = MateScore - 10000

// 探索中のスコア(ply=ルートからの距離に依存する)を、
// 置換表に保存する ply 非依存のスコアに変換する。
func toTTScore(score, ply int) int32 {
	if score >= mateThreshold {
		return int32(score + ply)
	}
	if score <= -mateThreshold {
		return int32(score - ply)
	}
	return int32(score)
}

// 置換表から読み出した ply 非依存のスコアを、
// 現在の ply における探索スコアに変換する。
func fromTTScore(score int32, ply int) int {
	s := int(score)
	if s >= mateThreshold {
		return s - ply
	}
	if s <= -mateThreshold {
		return s + ply
	}
	return s
}

// probe: 一致するエントリがあり、かつ十分な深さで探索済みなら
// (score, best, true) を返す。窓に収まらない場合でも best(手順序のヒント)
// だけは可能なら返す。
func (tt *transTable) probe(key uint64, depth, ply, alpha, beta int) (int, *shogi.Action, bool) {

	e := &tt.entries[tt.index(key)]
	if e.key != key {
		return 0, nil, false
	}

	best := e.best

	if int(e.depth) < depth {
		return 0, best, false
	}

	score := fromTTScore(e.score, ply)

	switch e.flag {
	case ttExact:
		return score, best, true
	case ttLower:
		if score >= beta {
			return score, best, true
		}
	case ttUpper:
		if score <= alpha {
			return score, best, true
		}
	}

	return 0, best, false
}

// store: 探索結果を保存する。常時上書き。
func (tt *transTable) store(key uint64, depth, ply, score int, flag ttFlag, best *shogi.Action) {
	idx := tt.index(key)
	tt.entries[idx] = ttEntry{
		key:   key,
		depth: int8(depth),
		flag:  flag,
		score: toTTScore(score, ply),
		best:  best,
	}
}
