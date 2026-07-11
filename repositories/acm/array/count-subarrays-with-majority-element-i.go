package array

func countMajoritySubarrays(nums []int, target int) int {
	n := len(nums)
	sum := make([]int, n+1)
	for i := 0; i < n; i++ {
		cnt := 0
		num := nums[i]
		if num == target {
			cnt = 1
		}
		sum[i+1] = sum[i] + cnt
	}
	ans := 0
	for i := 0; i <= n; i++ {
		for j := i + 1; j <= n; j++ {
			cnt := sum[j] - sum[i]
			if cnt*2 > j-i {
				ans++
			}
		}
	}
	return ans
}
