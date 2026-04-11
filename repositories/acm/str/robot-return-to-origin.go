package str

func judgeCircle(moves string) bool {
	x, y := 0, 0
	n := len(moves)
	for i := 0; i < n; i++ {
		if moves[i] == 'U' {
			x++
			continue
		}
		if moves[i] == 'D' {
			x--
		}
		if moves[i] == 'L' {
			y--
		}
		if moves[i] == 'R' {
			y++
		}
	}
	if x == 0 && y == 0 {
		return true
	}
	return false
}
