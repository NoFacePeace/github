package math

// https://leetcode.cn/problems/type-of-triangle/?envType=daily-question&envId=2025-05-19

func triangleType(nums []int) string {

	a, b, c := nums[0], nums[1], nums[2]
	if a >= b+c || b >= a+c || c >= a+b {
		return "none"
	}
	if a == b && b == c {
		return "equilateral"
	}
	if a == b || b == c || a == c {
		return "isosceles"
	}

	return "scalene"
}
