package array

func differenceOfSum(nums []int) int {
	ele := 0
	num := 0
	for _, v := range nums {
		ele += v
		for v != 0 {
			num += v % 10
			v /= 10
		}
	}
	if ele > num {
		return ele - num
	}
	return num - ele
}
