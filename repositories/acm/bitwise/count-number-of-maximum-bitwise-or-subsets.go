package bitwise

func countMaxOrSubsets(nums []int) int {
	cnt := 0
	mx := 0
	n := len(nums)
	var dfs func(idx int, val int)
	dfs = func(idx int, val int) {

		if idx == n {
			if val > mx {
				mx = val
				cnt = 1
			} else if val == mx {
				cnt++
			}
			return
		}
		dfs(idx+1, val|nums[idx])
		dfs(idx+1, val)
	}
	dfs(0, 0)
	return cnt
}
