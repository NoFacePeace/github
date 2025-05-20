package dp

func numTilings(n int) int {
	ans := make([]int, n+1)
	sum := make([]int, n+1)
	mod := int(1e9) + 7
	for i := 1; i <= n; i++ {
		if i == 1 {
			ans[i] = 1
			sum[i] += ans[i]
			continue
		}
		if i == 2 {
			ans[i] = 2
			sum[i] += sum[i-1] + ans[i]
			continue
		}
		ans[i] = ans[i-1] + ans[i-2] + sum[i-3]*2 + 2
		ans[i] %= mod
		sum[i] += sum[i-1] + ans[i]
	}
	return ans[n]
}
