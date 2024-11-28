package bitwise

func singleNonDuplicate(nums []int) int {
	ans := -1
	for _, v := range nums {
		if ans == -1 {
			ans = v
		} else {
			ans ^= v
		}
	}
	return ans
}
