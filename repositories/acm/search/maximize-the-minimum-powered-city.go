package search

func maxPower(stations []int, r int, k int) int64 {
	ans := 0
	n := len(stations)
	compute := func() int {
		powers := make([]int, n)
		left, right := 0, 0
		for i := 0; i < n; i++ {
			// 先加自己
			powers[i] += stations[i]
			powers[i] += left
			if i-r >= 0 {
				left -= stations[i-r]
			}
			left += stations[i]
			if i-r >= 0 {
				right -= stations[i-r]

			}
			right += stations[i]
			if i-r >= 0 {
				powers[i-r] += right
			}

		}
		for i := n - r; i < n; i++ {
			right -= stations[i]
			powers[i] += right
		}
		p := powers[0]
		for i := 0; i < n; i++ {
			p = min(p, powers[i])
		}
		return p
	}
	var dfs func(idx, k int)
	dfs = func(idx, k int) {
		if k == 0 {
			val := compute()
			ans = max(ans, val)
			return
		}
		if idx == len(stations) {
			return
		}
		for i := 0; i <= k; i++ {
			stations[idx] += i
			dfs(idx+1, k-i)
			stations[idx] -= i
		}
	}
	dfs(0, k)
	return int64(ans)
}
