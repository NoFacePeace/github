package array

// https://leetcode.cn/problems/count-subarrays-of-length-three-with-a-condition/description/?envType=daily-question&envId=2025-04-27

func countSubarrays(nums []int) int {
	ans := 0
	n := len(nums)
	for i := 0; i < n-2; i++ {
		first := nums[i]
		second := nums[i+1]
		third := nums[i+2]
		if (first+third)*2 == second {
			ans++
		}
	}
	return ans
}
