package dp

import "math"

func minCost(maxTime int, edges [][]int, passingFees []int) int {
	n := len(passingFees)
	f := make([][]int, maxTime+1)
	for i := range f {
		f[i] = make([]int, n)
		for j := range f[i] {
			f[i][j] = math.MaxInt
		}
	}
	f[0][0] = passingFees[0]
	for t := 1; t <= maxTime; t++ {
		for _, edge := range edges {
			i, j, cost := edge[0], edge[1], edge[2]
			if cost <= t {
				if f[t-cost][j] != math.MaxInt {
					f[t][i] = min(f[t][i], f[t-cost][j]+passingFees[i])
				}
				if f[t-cost][i] != math.MaxInt {
					f[t][j] = min(f[t][j], f[t-cost][i]+passingFees[j])
				}
			}
		}
	}
	ans := math.MaxInt
	for t := 1; t <= maxTime; t++ {
		ans = min(ans, f[t][n-1])
	}
	if ans == math.MaxInt {
		return -1
	}
	return ans
}
