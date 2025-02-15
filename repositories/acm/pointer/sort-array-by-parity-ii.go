package pointer

// https://leetcode.cn/problems/sort-array-by-parity-ii/

func sortArrayByParityII(nums []int) []int {
	n := len(nums)
	l, r := 0, n-1
	for l < r {
		if l%2 == nums[l]%2 {
			l++
			r = n - 1
			continue
		}
		nums[l], nums[r] = nums[r], nums[l]
		r--
	}
	return nums
}
