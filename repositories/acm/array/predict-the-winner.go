package array

func predictTheWinner(nums []int) bool {
	var dfs func(nums []int) (int, int)
	dfs = func(nums []int) (int, int) {
		n := len(nums)
		if n == 0 {
			return 0, 0
		}
		if n == 1 {
			return nums[0], 0
		}
		one1, two1 := dfs(nums[1:])
		one2, two2 := dfs(nums[:n-1])
		if two1+nums[0] >= two2+nums[n-1] {
			return two1 + nums[0], one1
		}
		return two2 + nums[n-1], one2
	}
	one, two := dfs(nums)
	return one >= two
}
