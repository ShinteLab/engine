package shogi

// 千日手の判定結果
type RepetitionStatus int

const (
	RepetitionNone RepetitionStatus = iota
	//千日手(引き分け)
	RepetitionDraw
	//連続王手の千日手: 現局面の手番側が「王手を掛け続けた側」で、その側の負け
	RepetitionPerpetualLose
)

// 現局面の千日手判定を行う。
//
// 現局面のハッシュが局面履歴中に4回出現していれば千日手成立。
// さらに、直近の同一局面出現(3回目)から現局面(4回目)までの間、
// 現局面の手番側自身の指し手(その手を指した直後に相手が王手されている
// = checkHistory が true)がすべて王手だった場合は、連続王手の千日手として
// その手番側の負けとする。それ以外は通常の千日手(引き分け)。
//
// 判定は「現局面がちょうど4回目の出現」であることを前提にせず、
// 履歴全体から一致するハッシュを都度数える(過去に遡っての再判定はしないが、
// 4回以上溜まっていても正しく成立と判定できるようにするため)。
func (b *Board) Repetition() RepetitionStatus {

	current := b.hash

	var indices []int
	for i, h := range b.history {
		if h == current {
			indices = append(indices, i)
		}
	}

	if len(indices) < 4 {
		return RepetitionNone
	}

	n := len(indices)
	prev := indices[n-2]
	last := indices[n-1]

	//prev(3回目の出現)からlast(現局面, 4回目の出現)までの間、
	//現局面の手番側自身が指した手(prev+1, prev+3, ... , last-1 の位置に対応)が
	//すべて王手だったかを見る。
	//(prev, last はどちらも同じ手番側が指す局面のため、その間の偶数番目の
	// オフセットは相手側の手、奇数番目のオフセットが自分側の手になる)
	allSelfChecks := true
	for i := prev + 1; i < last; i += 2 {
		if !b.checkHistory[i] {
			allSelfChecks = false
			break
		}
	}

	if allSelfChecks {
		return RepetitionPerpetualLose
	}

	return RepetitionDraw
}
