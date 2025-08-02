package array

func countHillValley(nums []int) int {
	ans := 0
	n := len(nums)
	i := 0
	for i < n {
		if i == 0 {
			i++
			continue
		}
		if i == n-1 {
			i++
			continue
		}
		if nums[i] > nums[i-1] {
			for j := i + 1; j < n; j++ {
				if nums[i] == nums[j] {
					continue
				}
				if nums[i] > nums[j] {
					ans++
				}
				break
			}
		}
		if nums[i] < nums[i-1] {
			for j := i + 1; j < n; j++ {
				if nums[i] == nums[j] {
					continue
				}
				if nums[i] < nums[j] {
					ans++
				}
				break
			}
		}
		i++
	}
	return ans
}
