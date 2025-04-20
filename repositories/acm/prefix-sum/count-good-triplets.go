package prefixsum

// https://leetcode.cn/problems/count-good-triplets/solutions/371340/tong-ji-hao-san-yuan-zu-by-leetcode-solution/?envType=daily-question&envId=2025-04-14

func countGoodTriplets(arr []int, a int, b int, c int) int {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	ans := 0
	n := len(arr)
	sum := make([]int, 1001)
	for j := 0; j < n; j++ {
		for k := j + 1; k < n; k++ {
			if abs(arr[j]-arr[k]) <= b {
				lj, rj := arr[j]-a, arr[j]+a
				lk, rk := arr[k]-c, arr[k]+c
				l := max(0, max(lj, lk))
				r := min(1000, min(rj, rk))
				if l <= r {
					if l == 0 {
						ans += sum[r]
					} else {
						ans += sum[r] - sum[l-1]
					}
				}
			}
		}
		for k := arr[j]; k <= 1000; k++ {
			sum[k]++
		}
	}
	return ans
}
