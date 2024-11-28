package array

func winningPlayerCount(n int, pick [][]int) int {
	m := make([]map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = map[int]int{}
	}
	ans := 0
	visit := make([]bool, n)
	for _, v := range pick {
		p, c := v[0], v[1]
		m[p][c]++
		if m[p][c] > p && !visit[p] {
			ans++
			visit[p] = true
		}
	}
	return ans
}
