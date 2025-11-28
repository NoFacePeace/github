package greedy

func minCost(colors string, neededTime []int) int {
	arr := []byte(colors)
	n := len(arr)
	ans := 0
	l, r := 0, 0
	for r < n {
		if arr[l] == arr[r] {
			r++
			continue
		}
		cost := 0
		mx := neededTime[l]
		for i := l; i < r; i++ {
			cost += neededTime[i]
			mx = max(mx, neededTime[i])
		}
		ans += cost - mx
		l = r
	}
	cost := 0
	mx := neededTime[l]
	for i := l; i < r; i++ {
		cost += neededTime[i]
		mx = max(mx, neededTime[i])
	}
	ans += cost - mx
	return ans
}
