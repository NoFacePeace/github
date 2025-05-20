package bitwise

// https://leetcode.cn/problems/find-numbers-with-even-number-of-digits/?envType=daily-question&envId=2025-04-30

func findNumbers(nums []int) int {
	ans := 0
	for _, v := range nums {
		if v >= 100000 {
			ans++
			continue
		}
		if v >= 10000 {
			continue
		}
		if v >= 1000 {
			ans++
			continue
		}
		if v >= 100 {
			continue
		}
		if v >= 10 {
			ans++
			continue
		}
	}
	return ans
}
