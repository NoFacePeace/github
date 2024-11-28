package sort

func numFriendRequests(ages []int) int {
	const mx = 121
	var cnt, pre [mx]int
	for _, age := range ages {
		cnt[age]++
	}
	for i := 1; i < mx; i++ {
		pre[i] = pre[i-1] + cnt[i]
	}
	ans := 0
	for i := 15; i < mx; i++ {
		if cnt[i] > 0 {
			bound := i/2 + 8
			ans += cnt[i] * (pre[i] - pre[bound-1] - 1)
		}
	}
	return ans
}
