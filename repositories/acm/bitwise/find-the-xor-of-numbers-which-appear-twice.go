package bitwise

func duplicateNumbersXOR(nums []int) int {
	m := map[int]int{}
	for _, v := range nums {
		m[v]++
	}
	ans := 0
	for k, v := range m {
		if v == 1 {
			continue
		}
		if ans == 0 {
			ans = k
			continue
		}
		ans ^= k
	}
	return ans
}
