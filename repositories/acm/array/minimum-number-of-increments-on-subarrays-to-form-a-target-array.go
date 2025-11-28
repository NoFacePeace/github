package array

func minNumberOperations(target []int) int {
	var dfs func(arr []int, val int)
	ans := 0
	dfs = func(arr []int, val int) {
		if len(arr) == 0 {
			return
		}
		mn := arr[0]
		for _, v := range arr {
			mn = min(v, mn)
		}
		ans += mn - val + 1
		l := 0
		for k, v := range arr {
			if v != mn {
				continue
			}
			dfs(arr[l:k], mn+1)
			l = k + 1
		}
		dfs(arr[l:], mn+1)
	}
	dfs(target, 1)
	return ans
}
