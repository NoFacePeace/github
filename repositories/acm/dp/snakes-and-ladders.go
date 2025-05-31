package dp

func snakesAndLadders(board [][]int) int {
	n := len(board)
	dp := map[int]int{}
	i, j := n-1, 0
	idx := 1
	sign := 1
	pos := map[int][]int{}
	for idx <= n*n {
		pos[idx] = []int{i, j}
		idx++
		j += sign
		if j == n || j == -1 {
			i--
			if sign == 1 {
				j = n - 1
				sign = -1
			} else {
				j = 0
				sign = 1
			}
		}
	}
	visit := map[int]bool{}
	q := []int{1}
	for len(q) > 0 {
		num := q[0]
		q = q[1:]
		if visit[num] {
			continue
		}
		visit[num] = true
		for j := num + 1; j <= min(num+6, n*n); j++ {
			x, y := pos[j][0], pos[j][1]
			if board[x][y] == -1 {
				if dp[j] == 0 {
					dp[j] = dp[num] + 1
				} else {
					dp[j] = min(dp[j], dp[num]+1)
				}
				if !visit[j] {
					q = append(q, j)
				}
				continue
			}
			idx := board[x][y]
			if idx == 1 {
				continue
			}
			if dp[idx] == 0 {
				dp[idx] = dp[num] + 1
			} else {
				dp[idx] = min(dp[idx], dp[num]+1)
			}
			if !visit[idx] {
				q = append(q, idx)
			}
		}
	}
	if dp[n*n] == 0 {
		return -1
	}
	return dp[n*n]
}
