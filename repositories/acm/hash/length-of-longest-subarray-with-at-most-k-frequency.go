package hash

func maxSubarrayLength(nums []int, k int) int {
	n := len(nums)
	ans := 0
	i := 0
	m := map[int]int{}
	cnt := 0
	for j := 0; j < n; j++ {
		num := nums[j]
		m[num]++
		cnt++
		if m[num] <= k {
			ans = max(ans, cnt)
			continue
		}
		for i <= j {
			num := nums[i]
			m[num]--
			cnt--
			if num == nums[j] {
				break
			}
			i++
		}
	}
	return ans
}
