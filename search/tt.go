package search

import (
	"sync/atomic"

	"shinte/core/usi"
	"shinte/engine"
)

// 置換表のスコア種別
type ttFlag uint8

const (
	ttExact ttFlag = iota
	ttLower
	ttUpper
)

// Stage K: ロックフリー(xor-verify / lockless hashing)方式のエントリ。
// data に (score, depth, flag, 手の圧縮表現) を詰め、
// check = key ^ data として書き込む。読み出し側は data と check を
// (別々に)atomic に読み、check^data==key が成立する場合のみ有効な
// エントリとして扱う。並行書き込みにより data と check が別の書き込みに
// 由来する「tear」が起きても、その組み合わせが偶然 key と一致する確率は
// 無視できるほど小さいため、tear は単に miss として扱われる
// (ロックを取らずに複数 goroutine から安全に共有できる)。
type ttEntry struct {
	data  uint64
	check uint64
}

// 固定サイズ配列 + インデックスマスクによる置換表。
// map ではなくスライスを使うことで GC 負荷を避ける。
// 書き込みは常時上書き(replace)方式とする。
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

// --- data のビットレイアウト(63ビットまで使用、bit63は未使用) ---
//
//	bits  0-19 : 手の圧縮表現(encodeMove/decodeMoveを参照)
//	bit   20   : hasMove(手を保持しているか)
//	bits 21-22 : flag(ttExact/ttLower/ttUpper)
//	bits 23-30 : depth(0..255。実運用では数十以下)
//	bits 31-62 : score(int32のビットパターンをそのまま格納)
const (
	dataMoveShift    = 0
	dataMoveBits     = uint64(0xFFFFF) // 20 bits
	dataHasMoveShift = 20
	dataFlagShift    = 21
	dataFlagBits     = uint64(0x3)
	dataDepthShift   = 23
	dataDepthBits    = uint64(0xFF)
	dataScoreShift   = 31
	dataScoreBits    = uint64(0xFFFFFFFF)
)

func packData(depth int, flag ttFlag, score int32, moveBits uint32, hasMove bool) uint64 {
	d := uint64(moveBits) & dataMoveBits
	if hasMove {
		d |= 1 << dataHasMoveShift
	}
	d |= (uint64(flag) & dataFlagBits) << dataFlagShift
	d |= (uint64(uint8(depth)) & dataDepthBits) << dataDepthShift
	d |= (uint64(uint32(score)) & dataScoreBits) << dataScoreShift
	return d
}

func unpackData(d uint64) (depth int, flag ttFlag, score int32, moveBits uint32, hasMove bool) {
	moveBits = uint32(d & dataMoveBits)
	hasMove = (d>>dataHasMoveShift)&1 != 0
	flag = ttFlag((d >> dataFlagShift) & dataFlagBits)
	depth = int((d >> dataDepthShift) & dataDepthBits)
	score = int32(uint32((d >> dataScoreShift) & dataScoreBits))
	return
}

// --- 手の圧縮表現(20ビット) ---
//
//	bits 0-6  : from(sq 0..80。打ちの場合は未使用)
//	bits 7-13 : to(sq 0..80)
//	bit  14   : growth(成り)
//	bit  15   : hit(打ちか)
//	bits 16-19: hitType(打ちの場合の基本駒種、PieceType 0..13が収まる)
//
// Action の再構築は shogi.NewAction(文字列) 経由で行う(シンプルさ優先。
// action.go 側に非公開コンストラクタを追加する代わりに、Stage J までに
// 追加済みの公開アクセサ BeforeXY/AfterXY/Hit/HitType/Promotes だけで
// エンコードし、デコード時は USI 文字列を組み立てて NewAction に渡す)。
func encodeMove(a *shogi.Action) uint32 {

	var v uint32

	ax, ay := a.AfterXY()
	to := uint32(squareIndex(ax, ay))
	v |= to << 7

	if a.Hit() {
		v |= 1 << 15
		v |= uint32(a.HitType()) << 16
		return v
	}

	fx, fy := a.BeforeXY()
	from := uint32(squareIndex(fx, fy))
	v |= from
	if a.Promotes() {
		v |= 1 << 14
	}
	return v
}

func decodeMove(v uint32) *shogi.Action {

	to := int((v >> 7) & 0x7F)
	tx, ty := xyFromSquareIndex(to)
	toStr := usi.FormatSquare(tx, ty)

	hit := (v>>15)&1 != 0
	if hit {
		hitType := shogi.PieceType((v >> 16) & 0xF)
		return shogi.NewAction(hitType.Mark() + "*" + toStr)
	}

	from := int(v & 0x7F)
	fx, fy := xyFromSquareIndex(from)
	fromStr := usi.FormatSquare(fx, fy)

	str := fromStr + toStr
	if (v>>14)&1 != 0 {
		str += "+"
	}
	return shogi.NewAction(str)
}

func xyFromSquareIndex(sq int) (int, int) {
	return sq%9 + 1, sq/9 + 1
}

// probe: 一致するエントリがあり、かつ十分な深さで探索済みなら
// (score, best, true) を返す。窓に収まらない場合でも best(手順序のヒント)
// だけは可能なら返す。data/check は別々に atomic Load するため、
// 並行書き込みと競合した場合は check^data != key となり、
// 安全側(miss)に倒れる。
func (tt *transTable) probe(key uint64, depth, ply, alpha, beta int) (int, *shogi.Action, bool) {

	e := &tt.entries[tt.index(key)]

	data := atomic.LoadUint64(&e.data)
	check := atomic.LoadUint64(&e.check)

	if check^data != key {
		return 0, nil, false
	}

	edepth, flag, escore, moveBits, hasMove := unpackData(data)

	var best *shogi.Action
	if hasMove {
		best = decodeMove(moveBits)
	}

	if edepth < depth {
		return 0, best, false
	}

	score := fromTTScore(escore, ply)

	switch flag {
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

// store: 探索結果を保存する。常時上書き。data → check の順に atomic
// Store する(順序自体は tear 検出の正しさに影響しない。読み手は必ず
// check^data==key を検証するため、どちらの順で書いても未検証のまま
// 誤った値が使われることはない)。
func (tt *transTable) store(key uint64, depth, ply, score int, flag ttFlag, best *shogi.Action) {

	var moveBits uint32
	hasMove := best != nil
	if hasMove {
		moveBits = encodeMove(best)
	}

	data := packData(depth, flag, toTTScore(score, ply), moveBits, hasMove)
	check := key ^ data

	idx := tt.index(key)
	e := &tt.entries[idx]
	atomic.StoreUint64(&e.data, data)
	atomic.StoreUint64(&e.check, check)
}
