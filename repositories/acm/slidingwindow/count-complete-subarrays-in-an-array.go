package slidingwindow

func countCompleteSubarrays(nums []int) int {
	m := map[int]int{}
	for _, v := range nums {
		m[v]++
	}
	cnt := len(m)
	m = map[int]int{}
	l, r := 0, 0
	n := len(nums)
	ans := 0
	for r < n {
		num := nums[r]
		m[num]++
		if len(m) != cnt {
			r++
			continue
		}
		ans += n - r
		m[num]--
		if m[num] == 0 {
			delete(m, num)
		}
		if l == r {
			r++
			continue
		}
		num = nums[l]
		m[num]--
		if m[num] == 0 {
			delete(m, num)
		}
		l++
	}
	return ans
}
