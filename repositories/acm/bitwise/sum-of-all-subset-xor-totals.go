package bitwise

// https://leetcode.cn/problems/sum-of-all-subset-xor-totals/submissions/619093379/?envType=daily-question&envId=2025-04-05

func subsetXORSum(nums []int) int {
	res := 0
	n := len(nums)
	for _, num := range nums {
		res |= num
	}
	return res << (n - 1)
}
