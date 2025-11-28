package bitwise

func prefixesDivBy5(nums []int) []bool {
	mod := 0
	ans := []bool{}
	for _, v := range nums {
		mod *= 2
		mod += v
		mod %= 5
		if mod == 0 {
			ans = append(ans, true)
		} else {
			ans = append(ans, false)
		}
	}
	return ans
}
