package array

func minFlipsII(grid [][]int) int {
	m := len(grid)
	if m == 0 {
		return 0
	}
	n := len(grid[0])
	if n == 0 {
		return 0
	}
	ans := 0
	one := 0
	zero := 0
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if i >= (m+1)/2 || j >= (n+1)/2 {
				continue
			}
			cnt := 1
			diff := 0
			if i < m/2 {
				cnt++
				if grid[i][j] != grid[m-i-1][j] {
					diff++
				}
			}
			if j < n/2 {
				cnt++
				if grid[i][j] != grid[i][n-j-1] {
					diff++
				}
			}
			if i < m/2 && j < n/2 {
				cnt++
				if grid[i][j] != grid[m-i-1][n-j-1] {
					diff++
				}
			}
			if diff < 3 {
				ans += diff
			} else {
				ans++
			}
			if cnt == 4 {
				continue
			}
			if cnt == 1 {
				if grid[i][j] == 1 {
					one++
				}
				continue
			}
			if diff == 1 {
				zero += 2
				continue
			}
			if grid[i][j] == 1 {
				one += 2
			}
		}
	}
	one %= 4
	if one == 1 {
		return ans + 1
	}
	if one == 2 && zero == 0 {
		return ans + 2
	}
	if one == 3 && zero != 0 {
		return ans + 1
	}
	if one == 3 {
		return ans + 3
	}
	return ans
}
