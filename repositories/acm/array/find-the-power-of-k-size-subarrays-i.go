package array

func resultsArray(nums []int, k int) []int {
	left := 0
	right := 0
	ans := []int{}
	for i := 0; i < len(nums); i++ {
		if i == 0 {
			if k == 1 {
				ans = append(ans, nums[i])
			}
			continue
		}
		if nums[i]-1 == nums[right] {
			right++
			if right-left+1 > k {
				left++
			}
		} else {
			right++
			left = right
		}
		if right-left+1 == k {
			ans = append(ans, nums[i])
			continue
		}
		if i+1 >= k {
			ans = append(ans, -1)
		}
	}
	return ans
}
