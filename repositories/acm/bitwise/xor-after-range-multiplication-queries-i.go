package bitwise

func xorAfterQueries(nums []int, queries [][]int) int {
	if len(nums) == 0 {
		return 0
	}
	mod := int(1e9) + 7
	for _, query := range queries {
		l, r, k, v := query[0], query[1], query[2], query[3]
		for i := l; i <= r; i += k {
			nums[i] = (nums[i] * v) % mod
		}
	}
	ans := nums[0]
	for i := 1; i < len(nums); i++ {
		ans ^= nums[i]
	}
	return ans
}
