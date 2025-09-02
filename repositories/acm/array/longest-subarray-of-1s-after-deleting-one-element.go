package array

func longestSubarray(nums []int) int {
	left := 0
	right := -1
	ans := 0
	for _, v := range nums {
		if v == 1 {
			if right != -1 {
				right++
			} else {
				left++
			}
			continue
		}
		if right == -1 {
			right++
			continue
		}
		if left+right > ans {
			ans = left + right
		}
		left = right
		right = 0
	}
	if left+right > ans {
		ans = left + right
	}
	return ans
}
