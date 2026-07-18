package shogi

// 盤面位置情報
type Pos [2]uint8

var PosNone = newPos(0, 0)

func newPos(x, y int) Pos {
	var p Pos
	p[0] = uint8(x)
	p[1] = uint8(y)
	return p
}

func parsePos(buf string) Pos {

	if len(buf) != 2 {
		return PosNone
	}

	//49 - 57 1-9
	x := buf[0] - 48
	//座標系とは逆
	x = 10 - x

	//97 - 105 a-i
	y := buf[1] - 96
	return newPos(int(x), int(y))
}

func (p Pos) XY() (int, int) {
	return int(p[0]), int(p[1])
}

func (p Pos) String() string {

	//送信系は逆になる
	x := byte((10 - p[0]) + 48)

	y := byte(p[1] + 96)
	return string(x) + string(y)
}
