package array

import "sort"

func combinationSum2(candidates []int, target int) [][]int {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i] < candidates[j]
	})
	var dfs func(idx, sum int, arr []int)
	ans := [][]int{}
	dfs = func(idx, sum int, arr []int) {
		if idx >= len(candidates) {
			return
		}
		val := candidates[idx]
		if val+sum > target {
			return
		}
		if val+sum == target {
			arr = append(arr, val)
			tmp := append([]int{}, arr...)
			ans = append(ans, tmp)
			return
		}
		arr = append(arr, val)
		dfs(idx+1, sum+val, arr)
		arr = arr[:len(arr)-1]
		for idx+1 < len(candidates) {
			if val != candidates[idx+1] {
				dfs(idx+1, sum, arr)
				break
			}
			idx = idx + 1
		}
	}
	dfs(0, 0, []int{})
	return ans
}
