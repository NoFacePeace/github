package array

func numSpecial(mat [][]int) int {
	m := len(mat)
	if m == 0 {
		return 0
	}
	n := len(mat[0])
	if n == 0 {
		return 0
	}
	rows := make([]int, m)
	cols := make([]int, n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			rows[i] += mat[i][j]
			cols[j] += mat[i][j]
		}
	}
	cnt := 0
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if rows[i] == 1 && cols[j] == 1 && mat[i][j] == 1 {
				cnt++
			}
		}
	}
	return cnt
}
