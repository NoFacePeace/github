package hash

// https://leetcode.cn/problems/count-number-of-bad-pairs/?envType=daily-question&envId=2025-04-18

func countBadPairs(nums []int) int64 {
	ans := 0
	n := len(nums)
	m := map[int]int{}
	for i := 0; i < n; i++ {
		num := nums[i]
		cnt := m[num-i]
		ans += i - cnt
		m[num-i]++
	}
	return int64(ans)
}
