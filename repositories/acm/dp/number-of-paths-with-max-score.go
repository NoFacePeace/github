package dp

func pathsWithMaxScore(board []string) []int {
	m := len(board)
	if m == 0 {
		return []int{0, 0}
	}
	n := len(board[0])
	if n == 0 {
		return []int{0, 0}
	}
	dp := make([][]int, m)
	score := make([][]int, m)
	for i := range dp {
		dp[i] = make([]int, n)
		score[i] = make([]int, n)
	}
	mod := int(1e9) + 7
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if board[i][j] == 'X' {
				continue
			}
			s := int(board[i][j] - '0')
			if i == m-1 && j == n-1 {
				score[i][j] = 0
				dp[i][j] = 1
				continue
			}
			if i == 0 && j == 0 {
				s = 0
			}
			if i == m-1 {
				if dp[i][j+1] == 0 {
					continue
				}
				dp[i][j] = dp[i][j+1]
				score[i][j] = score[i][j+1] + s
				continue
			}
			if j == n-1 {
				if dp[i+1][j] == 0 {
					continue
				}
				dp[i][j] = dp[i+1][j]
				score[i][j] = score[i+1][j] + s
				continue
			}
			if dp[i+1][j] != 0 {
				dp[i][j] = dp[i+1][j]
				score[i][j] = score[i+1][j] + s
			}
			if dp[i][j+1] != 0 {
				if score[i][j] < score[i][j+1]+s {
					dp[i][j] = dp[i][j+1]
					score[i][j] = score[i][j+1] + s
				} else if score[i][j] == score[i][j+1]+s {
					dp[i][j] += dp[i][j+1]
				}
			}
			dp[i][j] %= mod
			if dp[i+1][j+1] != 0 {
				if score[i][j] < score[i+1][j+1]+s {
					dp[i][j] = dp[i+1][j+1]
					score[i][j] = score[i+1][j+1] + s
				} else if score[i][j] == score[i+1][j+1]+s {
					dp[i][j] += dp[i+1][j+1]
				}
			}
			dp[i][j] %= mod
		}
	}
	return []int{score[0][0], dp[0][0]}
}
