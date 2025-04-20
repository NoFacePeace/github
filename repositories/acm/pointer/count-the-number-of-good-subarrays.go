package pointer

// https://leetcode.cn/problems/count-the-number-of-good-subarrays/

func countGood(nums []int, k int) int64 {
	n := len(nums)
	m := map[int]int{}
	ans := 0
	pair := 0
	l := 0
	for i := 0; i < n; i++ {
		num := nums[i]
		m[num]++
		pair += m[num] - 1
		if pair < k {
			continue
		}
		ans += n - i
		for l <= i {
			num := nums[l]
			m[num]--
			pair -= m[num]
			l++
			if pair < k {
				break
			}
			ans += n - i
		}
	}
	return int64(ans)
}
