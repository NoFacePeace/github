package search

import "math"

func minScore(n int, roads [][]int) int {
	ad := make([]map[int]int, n+1)
	for _, road := range roads {
		a, b, d := road[0], road[1], road[2]
		if ad[a] == nil {
			ad[a] = map[int]int{}
		}
		if ad[b] == nil {
			ad[b] = map[int]int{}
		}
		ad[a][b] = d
		ad[b][a] = d
	}
	q := []int{1}
	visited := map[int]bool{}
	ans := math.MaxInt
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		if visited[u] {
			continue
		}
		visited[u] = true
		for k, v := range ad[u] {
			ans = min(ans, v)
			q = append(q, k)
		}
	}
	return ans
}
