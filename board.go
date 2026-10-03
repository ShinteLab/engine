package shogi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ShinteLab/core/sfen"
	"golang.org/x/xerrors"
)

// 盤面
type Board struct {
	turn TurnType
	num  int

	camps [2]*CampBoard

	hash uint64

	//局面履歴(各手適用後のHash。初期局面が1件目)
	history []uint64
	//history[i] は「history[i]の局面に至る直前の手を指した側が、
	//その手で(次に指す側に)王手を掛けたか」。初期局面(index 0)は false。
	checkHistory []bool
}

const (
	StartPos     = "startpos"
	SFEN         = "sfen"
	StartPosSFEN = "lnsgkgsnl/1r5b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL"
	SFENLine     = "/"
	Moves        = "moves"
	None         = "-" //手駒
)

func initBoard() *Board {
	var b Board
	b.turn = TurnBlack
	b.num = 1
	b.camps[0] = NewCampBoard(TurnBlack)
	b.camps[1] = NewCampBoard(TurnWhite)
	//相手を設定
	b.camps[0].setEnemy(b.camps[1])
	return &b
}

func NewBoard(line string) (*Board, error) {

	//後手で相手が指した時
	//[startpos moves 8g8f]"
	//上手で指す時
	//[sfen lnsgkgsnl/7b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL w - 1]"
	//並列処理を意識した作りにする
	//[lnsgkgsnl/7b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/LNSGKGSNL w - 1 moves 8c8d]"
	//position startpos moves 6g6f 3c3d 5g5f
	b := initBoard()

	//strings.Fields は連続・前後の余分な空白を無視してトークン化する
	//(strings.Split(line," ")だと末尾の空白等が空文字列トークンを生み、
	// 以降のインデックス処理でパニックしうるため使わない)。
	args := strings.Fields(line)
	leng := len(args)
	if leng == 0 {
		return nil, fmt.Errorf("sfen parse error[empty]")
	}
	sfen := ""

	idx := 0
	t := args[idx]

	if t == StartPos {
		sfen = StartPosSFEN
	} else if t == SFEN {
		idx++
		sfen = args[idx]
	} else {
		return nil, fmt.Errorf("sfen parse error[%s]", line)
	}

	err := b.parse(sfen)
	if err != nil {
		return nil, err
	}

	idx++

	for {

		if idx >= leng {
			break
		}

		next := args[idx]
		if next != Moves {

			t := next
			h := args[idx+1]
			n := args[idx+2]

			err = b.SetStatus(t, h, n)
			if err != nil {
				return nil, xerrors.Errorf("SetStatus() error: %w", err)
			}
			idx = idx + 3
		} else {
			idx++

			//手を適用する前(=初期局面)のハッシュをhistoryの1件目として記録する
			b.hash = b.computeHash()
			b.history = append(b.history, b.hash)
			b.checkHistory = append(b.checkHistory, false)

			ms := args[idx:]
			//動作させる
			for _, mov := range ms {
				a := NewAction(mov)
				if a == nil {
					//パース不能な手(nilガード)
					return nil, fmt.Errorf("invalid move string[%s]", mov)
				}
				if !b.Action(a) {
					return nil, fmt.Errorf("Action() failed for move[%s]", mov)
				}
				idx++
			}
			break
		}
	}

	//movesが無かった場合はここで初期局面のハッシュ・履歴を記録する
	//(movesがあった場合は上のブロックとb.Action()経由で既に記録済み)
	if len(b.history) == 0 {
		b.hash = b.computeHash()
		b.history = append(b.history, b.hash)
		b.checkHistory = append(b.checkHistory, false)
	}

	return b, nil
}

// 盤面(camps)のみを複製する共通部。hash・履歴は複製しない。
func (b *Board) copyCore() *Board {

	var dst Board
	dst.turn = b.turn

	dst.camps[0] = b.camps[0].copy()
	dst.camps[1] = b.camps[1].copy()
	dst.camps[0].setEnemy(dst.camps[1])

	return &dst
}

func (b *Board) Copy() *Board {

	dst := b.copyCore()
	dst.hash = b.hash

	dst.history = make([]uint64, len(b.history))
	copy(dst.history, b.history)
	dst.checkHistory = make([]bool, len(b.checkHistory))
	copy(dst.checkHistory, b.checkHistory)

	return dst
}

// 軽量版 Copy: hash・履歴を複製しない。
// legalCandidate() 内の合法性フィルタ(IsCheck 判定用の使い捨てコピー)専用。
// このコピーに対して Repetition()/Hash() を呼び出してはならない
// (history が nil のため Repetition() は常に RepetitionNone を返す安全側の
// 挙動にはなるが、意味のある結果にはならない)。
func (b *Board) copyLite() *Board {
	return b.copyCore()
}

// 座標位置にピースを配置
func (b *Board) set(x, y int, p *Piece) error {
	p.set(x, y)
	b.camps[p.turn.Index()].set(x, y, p.typ)
	return nil
}

// SFEN の盤面部分を解析して駒を配置する。盤面文字列の解析仕様は
// core/sfen に集約している。SFEN 記述順の rank/file(0..8)を内部座標
// (x=file+1, y=rank+1)へ写す。
func (b *Board) parse(sfenBoard string) error {
	return sfen.ParseBoard(sfenBoard, func(rank, file, base int, black, promoted bool) {
		typ := PieceType(base)
		if promoted {
			typ = typ.Growth()
		}
		turn := TurnBlack
		if !black {
			turn = TurnWhite
		}
		b.set(file+1, rank+1, NewPieceFromType(typ, turn))
	})
}

// ターン 持ち駒 ターン数の文字列
func (b *Board) SetStatus(t string, h string, n string) error {
	b.turn = TurnType(t)
	b.setHas(h)
	var err error
	b.num, err = strconv.Atoi(n)
	if err != nil {
		logger().Error(fmt.Sprintf("Turn number Cast error:[%s]", n))
		b.num = 1
	}
	return nil
}

// 持ちゴマを文字列から設定
func (b *Board) setHas(h string) bool {
	for idx := 0; idx < len(h); idx++ {

		numBuf := ""
		var c byte
		for {
			c = h[idx]
			if c >= '0' && c <= '9' {
				numBuf += string(c)
				idx++
			} else {
				c = h[idx]
				if numBuf == "" {
					numBuf = "1"
				}
				break
			}
		}

		if c == '-' {
			break
		}

		num, _ := strconv.Atoi(numBuf)

		p := NewPieceFromString(string(c))
		for idx := 1; idx <= num; idx++ {
			b.camps[p.turn.Index()].has.add(p.typ)
		}
	}
	return true
}

// 現在の手番
func (b *Board) Turn() TurnType {
	return b.turn
}

// 手番 t 側の盤上駒価値 + 持駒価値の合計(探索の評価関数向け)。
// 王も Value() に含めた単純な差分評価用途(詰み検出は探索側で行う)。
func (b *Board) Material(t TurnType) int {

	camp := b.camps[t.Index()]
	sum := 0

	for pt := 0; pt < len(camp.typBoards); pt++ {
		bit := &camp.typBoards[pt]
		if bit.isZero() {
			continue
		}
		v := PieceType(pt).Value()
		bit.forEach(func(sq int) {
			sum += v
		})
	}

	for pt, cnt := range camp.has {
		if cnt > 0 {
			sum += PieceType(pt).Value() * cnt
		}
	}

	return sum
}

// 現在の手番が王手されているか(IsCheck(Turn()) の別名)
func (b *Board) InCheck() bool {
	return b.IsCheck(b.turn)
}

// 手番 t 側の各駒の位置を列挙する(PST等の位置評価向け)。
// sq は squareOf(x,y) と同じマス番号(0..80, (y-1)*9+(x-1))。
func (b *Board) PieceSquares(t TurnType, fn func(pt PieceType, sq int)) {

	camp := b.camps[t.Index()]

	for pt := 0; pt < len(camp.typBoards); pt++ {
		bit := &camp.typBoards[pt]
		if bit.isZero() {
			continue
		}
		pieceType := PieceType(pt)
		bit.forEach(func(sq int) {
			fn(pieceType, sq)
		})
	}
}

// 盤面への適用のみを行う共通部。手番・手数の更新のみで、
// hash再計算・履歴追記は行わない。
func (b *Board) actionCore(a *Action) bool {

	rtn := b.camps[b.turn.Index()].action(a)
	if !rtn {
		return false
	}
	b.turn = b.turn.Next()
	b.num++

	return true
}

// 動作させる
func (b *Board) Action(a *Action) bool {

	if !b.actionCore(a) {
		return false
	}

	//盤面変更後のハッシュを再計算する。
	//差分更新ではなく全再計算だが、typBoards の forEach で盤上駒数十個分の
	//XOR で済むため Action あたり数百ns程度に収まり、探索の Copy+Candidate
	//コスト(µs級)に比べ十分小さい。真の差分更新は将来最適化とする。
	b.hash = b.computeHash()

	//局面履歴に追記する。checkHistory は「今の手番側(次に指す側)が
	//直前の手によって王手を掛けられているか」、つまり直前の手を指した側が
	//王手を掛けたかどうかを表す。
	b.history = append(b.history, b.hash)
	b.checkHistory = append(b.checkHistory, b.IsCheck(b.turn))

	return true
}

// 軽量版 Action: hash再計算・履歴追記をスキップする。
// legalCandidate() 内の合法性フィルタ(IsCheck 判定用の使い捨てコピー)専用。
func (b *Board) actionLite(a *Action) bool {
	return b.actionCore(a)
}

// DoMove/UndoMove の復元に必要な情報。
type Undo struct {
	action *Action

	movedType    PieceType
	placedType   PieceType
	capturedType PieceType

	prevHash       uint64
	prevTurn       TurnType
	prevNum        int
	prevHistoryLen int
}

// 手を適用し、UndoMove で完全に元へ戻すための Undo を返す(単一 Board を
// 掘り下げる探索向け)。Zobrist ハッシュは全再計算ではなく差分更新する。
// 失敗時は (Undo{}, false) を返し、盤面は変更しない
// (actionInfo が false を返した時点でCampBoard側は既に一部変更している
//  可能性があるが、これは既存 Action()/actionCore と同じ制約であり
//  Stage I で新たに導入した挙動ではない)。
func (b *Board) DoMove(a *Action) (Undo, bool) {

	var u Undo
	u.action = a
	u.prevHash = b.hash
	u.prevTurn = b.turn
	u.prevNum = b.num
	u.prevHistoryLen = len(b.history)

	moverTurn := b.turn
	own := b.camps[moverTurn.Index()]

	info, ok := own.actionInfo(a)
	if !ok {
		return Undo{}, false
	}
	u.movedType = info.movedType
	u.placedType = info.placedType
	u.capturedType = info.capturedType

	b.turn = b.turn.Next()
	b.num++

	moverIdx := moverTurn.Index()
	enemyIdx := b.turn.Index()

	ax, ay := a.after.XY()
	destSq := squareOf(ax, ay)

	h := b.hash

	if a.Hit() {
		base := info.placedType.Base()
		newCount := own.has[base]
		oldCount := newCount + 1
		h ^= handHashComponent(moverIdx, base, oldCount)
		h ^= handHashComponent(moverIdx, base, newCount)

		h ^= zobristPiece[moverIdx][info.placedType][destSq]
	} else {
		bx, by := a.before.XY()
		srcSq := squareOf(bx, by)

		h ^= zobristPiece[moverIdx][info.movedType][srcSq]
		h ^= zobristPiece[moverIdx][info.placedType][destSq]
	}

	if info.capturedType != PieceTypeNotFound {
		h ^= zobristPiece[enemyIdx][info.capturedType][destSq]

		base := info.capturedType.Base()
		newCount := own.has[base]
		oldCount := newCount - 1
		h ^= handHashComponent(moverIdx, base, oldCount)
		h ^= handHashComponent(moverIdx, base, newCount)
	}

	//手番は必ず反転するので無条件にXORする(computeHash()の
	//「Whiteならzobristturnを立てる」と同じ効果になるトグル)
	h ^= zobristTurn

	b.hash = h

	b.history = append(b.history, b.hash)
	b.checkHistory = append(b.checkHistory, b.IsCheck(b.turn))

	return u, true
}

// DoMove で適用した手を完全に元へ戻す。
func (b *Board) UndoMove(u Undo) {

	a := u.action
	mover := b.camps[u.prevTurn.Index()]

	ax, ay := a.after.XY()

	//移動先に置かれた駒(placedType)を取り除く
	mover.typBoards[u.placedType].clear(ax, ay)
	mover.board.clear(ax, ay)

	if u.capturedType != PieceTypeNotFound {
		//捕獲していた敵駒を盤に戻す
		enemy := mover.enemy
		enemy.typBoards[u.capturedType].set(ax, ay)
		enemy.board.set(ax, ay)
		//持駒から減らす(has.removeはBase()を使うのでcapturedTypeが
		//成り駒でも正しく基本種のカウントから減る)
		mover.has.remove(u.capturedType)
	}

	if a.Hit() {
		//打った駒を持駒に戻す
		mover.has.add(u.placedType)
	} else {
		//移動元に駒(movedType, 成り前)を戻す
		bx, by := a.before.XY()
		mover.typBoards[u.movedType].set(bx, by)
		mover.board.set(bx, by)
	}

	b.turn = u.prevTurn
	b.num = u.prevNum
	b.hash = u.prevHash

	//historyはappendしているだけなので長さを戻すだけでよい。
	//スライスの容量は共有されたままだが、再DoMoveで上書きされるだけなので
	//問題ない(このBoardの外に別途Copy()したhistoryはmake+copyで
	//独立しているため影響を受けない)。
	b.history = b.history[:u.prevHistoryLen]
	b.checkHistory = b.checkHistory[:u.prevHistoryLen]
}

// 検証付きで手を適用する(GUI 等、信頼できない入力向け)。
// a が nil の場合や、現局面の合法手(Candidate())に含まれない場合は
// エラーを返して盤面を変更しない。ホットパス(探索内部)では検証コストを
// 避けるため引き続き無検証の Action() を使うこと。
func (b *Board) Move(a *Action) error {

	if a == nil {
		return fmt.Errorf("Move() error: action is nil")
	}

	s := a.String()
	legal := false
	for _, c := range b.Candidate() {
		if c.String() == s {
			legal = true
			break
		}
	}
	if !legal {
		return fmt.Errorf("Move() error: illegal move[%s]", s)
	}

	if !b.Action(a) {
		return fmt.Errorf("Move() error: Action() failed for [%s]", s)
	}

	return nil
}

// 現在の局面のハッシュ値(Zobrist ハッシュ)
func (b *Board) Hash() uint64 {
	return b.hash
}

// 局面全体からハッシュ値を再計算する
func (b *Board) computeHash() uint64 {

	var h uint64

	for ti := 0; ti < 2; ti++ {
		camp := b.camps[ti]

		for pt := 0; pt < len(camp.typBoards); pt++ {
			bit := &camp.typBoards[pt]
			if bit.isZero() {
				continue
			}
			bit.forEach(func(sq int) {
				h ^= zobristPiece[ti][pt][sq]
			})
		}

		for pt, cnt := range camp.has {
			if cnt <= 0 {
				continue
			}
			c := cnt
			if c >= len(zobristHand[ti][pt]) {
				c = len(zobristHand[ti][pt]) - 1
			}
			h ^= zobristHand[ti][pt][c]
		}
	}

	if b.turn == TurnWhite {
		h ^= zobristTurn
	}

	return h
}

// 手番 t の玉が相手の利きに入っているか(王手されているか)。
// 玉が盤上にない場合(テスト用局面等)は false を返す。
func (b *Board) IsCheck(t TurnType) bool {

	camp := b.camps[t.Index()]
	king := camp.typBoards[King]
	if king.isZero() {
		return false
	}

	enemy := camp.enemy
	occ := camp.board
	occ.or(&enemy.board)

	attackers := enemy.attackAll(&occ)
	attackers.and(&king)
	return !attackers.isZero()
}

// 擬似合法手(王手放置・自殺手・打ち歩詰めを含みうる)。
// テスト・内部使用向け。
func (b *Board) pseudoCandidate() []*Action {
	now := b.camps[b.turn.Index()]
	return now.pseudoCandidate()
}

// 合法手化の本体(Stage H以前の実装。copyLite+actionLite+IsCheckを
// 擬似合法手1手ごとに行う)。legal_diff_test.go の参照実装として残す。
// checkUchifuzume が false の場合は打ち歩詰め判定を省略する
// (打ち歩詰め判定が相手の合法手数を数えるために自身を再帰呼び出しするので、
//
//	無限再帰を避けるためこの1段だけ判定を止める)。
func (b *Board) legalCandidateSlow(checkUchifuzume bool) []*Action {

	moves := b.pseudoCandidate()

	legal := make([]*Action, 0, len(moves))
	for _, a := range moves {

		//IsCheck判定用の使い捨てコピーなのでhash・履歴は不要(copyLite/actionLite)
		nb := b.copyLite()
		if !nb.actionLite(a) {
			//現状は発生しないはずだがガード
			continue
		}

		//動かした側(自分)の玉が相手の利きに入る手(王手放置・自殺手)は除外
		if nb.IsCheck(b.turn) {
			continue
		}

		//打ち歩詰めの除外
		if checkUchifuzume && a.Hit() && a.HitType() == Pawn {
			if nb.IsCheck(nb.turn) {
				//打った歩によって王手になっている場合、相手に合法手が
				//1つも無ければ打ち歩詰めなので除外する。
				//ここでの合法手判定は打ち歩詰め判定を含めない(無限再帰防止)。
				if len(nb.legalCandidateSlow(false)) == 0 {
					continue
				}
			}
		}

		legal = append(legal, a)
	}
	return legal
}

// 合法手化の本体(Stage H: ピン検出方式)。局面につき1回 legalInfo を
// 計算し、各擬似合法手を軽量判定でふるいにかける。
// 玉が盤上に無い(テスト用局面等)場合は legalCandidateSlow にフォールバックする。
func (b *Board) legalCandidate(checkUchifuzume bool) []*Action {

	own := b.camps[b.turn.Index()]
	if own.typBoards[King].isZero() {
		return b.legalCandidateSlow(checkUchifuzume)
	}

	moves := b.pseudoCandidate()
	info := b.computeLegalInfo()

	legal := make([]*Action, 0, len(moves))
	for _, a := range moves {

		if !info.isLegal(a) {
			continue
		}

		//打ち歩詰めの除外(頻度が低いのでcopyLite経路を維持する)
		if checkUchifuzume && a.Hit() && a.HitType() == Pawn {
			nb := b.copyLite()
			if !nb.actionLite(a) {
				continue
			}
			if nb.IsCheck(nb.turn) {
				if len(nb.legalCandidate(false)) == 0 {
					continue
				}
			}
		}

		legal = append(legal, a)
	}
	return legal
}

// 動かせる箇所を取得(合法手のみ)
func (b *Board) Candidate() []*Action {
	return b.legalCandidate(true)
}

// 座標にある駒を取得
func (b *Board) Get(x, y int) *Piece {
	p := b.camps[0].Get(x, y)
	if !p.Empty() {
		return p
	}
	return b.camps[1].Get(x, y)
}

// 盤面表示
func (b *Board) GoString() string {

	var builder strings.Builder

	black := "*"
	white := " "
	if b.turn == TurnWhite {
		black = " "
		white = "*"
	}

	builder.WriteString(fmt.Sprintf(" |  9  8  7  6  5  4  3  2  1  |\n"))
	builder.WriteString(fmt.Sprintf("-------------------------------|\n"))

	for y := 1; y <= 9; y++ {
		builder.WriteString(fmt.Sprintf("%d|  ", y))
		for x := 1; x <= 9; x++ {
			p := b.Get(x, y)
			builder.WriteString(fmt.Sprintf("%-3s", p.Mark()))
		}
		builder.WriteString("|\n")
	}
	builder.WriteString(fmt.Sprintf("-------------------------------|\n"))
	builder.WriteString(fmt.Sprintf("%sBlack:%v\n", black, b.camps[0].has))
	builder.WriteString(fmt.Sprintf("%sWhite:%v\n", white, b.camps[1].has))
	return builder.String()
}
