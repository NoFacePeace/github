package prefixsum

// https://leetcode.cn/problems/count-of-interesting-subarrays/solutions/3647292/tong-ji-qu-wei-zi-shu-zu-de-shu-mu-by-le-968z/?envType=daily-question&envId=2025-04-25

func countInterestingSubarrays(nums []int, modulo int, k int) int64 {
	n := len(nums)
	cnt := map[int]int{}
	ans := 0
	prefix := 0
	cnt[0] = 1
	for i := 0; i < n; i++ {
		num := nums[i]
		if num%modulo == k {
			prefix++
		}
		ans += cnt[(prefix-k+modulo)%modulo]
		cnt[prefix%modulo]++
	}
	return int64(ans)
}
