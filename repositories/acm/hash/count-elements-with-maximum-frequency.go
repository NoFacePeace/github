package hash

func maxFrequencyElements(nums []int) int {
	m := map[int]int{}
	mx := 0
	for _, v := range nums {
		m[v]++
		mx = max(mx, m[v])
	}
	ans := 0
	for _, v := range m {
		if v == mx {
			ans += v
		}
	}
	return ans
}
