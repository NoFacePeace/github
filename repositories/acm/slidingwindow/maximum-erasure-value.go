package slidingwindow

func maximumUniqueSubarray(nums []int) int {
	l := 0
	m := map[int]bool{}
	ans := 0
	sum := 0
	for _, v := range nums {
		if !m[v] {
			sum += v
			ans = max(ans, sum)
			m[v] = true
			continue
		}
		num := nums[l]
		for num != v {
			m[num] = false
			sum -= num
			l++
			num = nums[l]
		}
		l++
	}
	return ans
}
