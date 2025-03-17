package bitwise

// https://leetcode.cn/problems/count-the-number-of-beautiful-subarrays/description/

func beautifulSubarrays(nums []int) int64 {
	cnt := map[int]int{}
	mask := 0
	ans := 0
	cnt[0] = 1
	for _, x := range nums {
		mask ^= x
		ans += cnt[mask]
		cnt[mask]++
	}
	return int64(ans)
}
