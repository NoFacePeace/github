package dp

// https://leetcode.cn/problems/partition-equal-subset-sum/submissions/619755671/?envType=daily-question&envId=2025-04-07

func canPartition(nums []int) bool {
	n := len(nums)
	if n < 2 {
		return false
	}
	sum, max := 0, 0
	for _, v := range nums {
		sum += v
		if v > max {
			max = v
		}
	}
	if sum%2 != 0 {
		return false
	}
	target := sum / 2
	if max > target {
		return false
	}
	dp := make([]bool, target+1)
	dp[0] = true
	for i := 0; i < n; i++ {
		v := nums[i]
		for j := target; j >= v; j-- {
			dp[j] = dp[j] || dp[j-v]
		}
	}
	return dp[target]
}
