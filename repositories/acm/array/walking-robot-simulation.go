package array

func robotSim(commands []int, obstacles [][]int) int {
	m := map[int]map[int]bool{}
	for _, v := range obstacles {
		x, y := v[0], v[1]
		if m[x] == nil {
			m[x] = map[int]bool{}
		}
		m[x][y] = true
	}
	x, y := 0, 0
	dx, dy := 0, 1
	n := len(commands)
	ans := 0
	for i := 0; i < n; i++ {
		if commands[i] == -1 {
			if dy == 1 {
				dx = 1
				dy = 0
				continue
			}
			if dy == -1 {
				dx = -1
				dy = 0
				continue
			}
			if dx == 1 {
				dy = -1
				dx = 0
				continue
			}
			dy = 1
			dx = 0
			continue
		}
		if commands[i] == -2 {
			if dy == 1 {
				dx = -1
				dy = 0
				continue
			}
			if dy == -1 {
				dx = 1
				dy = 0
				continue
			}
			if dx == 1 {
				dy = 1
				dx = 0
				continue
			}
			dy = -1
			dx = 0
			continue
		}
		step := commands[i]
		for i := 0; i < step; i++ {
			nx := x + dx*1
			ny := y + dy*1
			if m[nx][ny] {
				break
			}
			x = nx
			y = ny
		}
		ans = max(ans, x*x+y*y)
	}
	return ans
}
