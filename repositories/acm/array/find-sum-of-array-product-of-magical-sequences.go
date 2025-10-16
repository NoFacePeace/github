package array

func magicalSum(m int, k int, nums []int) int {
	ans := 0
	n := len(nums)
	mod := int(1e9) + 7
	var dfs func(idx, cnt, val int)
	dfs = func(idx, cnt, val int) {
		if cnt == k {
			ans += val
			ans %= mod
			return
		}
		if idx == n {
			return
		}
		dfs(idx+1, cnt+1, val*nums[idx])
		dfs(idx+1, cnt, val)
	}
	dfs(0, 0, 1)
	return ans
}
