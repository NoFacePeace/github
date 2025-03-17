package backtrack

// https://leetcode.cn/problems/the-number-of-beautiful-subsets/

func beautifulSubsets(nums []int, k int) int {
	ans := 0
	cnt := make(map[int]int)
	var dfs func(i int)
	dfs = func(i int) {
		if i == len(nums) {
			ans++
			return
		}
		dfs(i + 1)
		if cnt[nums[i]-k] == 0 && cnt[nums[i]+k] == 0 {
			cnt[nums[i]]++
			dfs(i + 1)
			cnt[nums[i]]--
		}
	}
	dfs(0)
	return ans - 1
}
