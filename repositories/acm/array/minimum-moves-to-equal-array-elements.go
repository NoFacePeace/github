package array

func minMoves(nums []int) int {
	mn := nums[0]
	for _, v := range nums {
		mn = min(mn, v)
	}
	ans := 0
	for _, v := range nums {
		ans += v - mn
	}
	return ans
}
