package greedy

// https://leetcode.cn/problems/maximum-value-of-an-ordered-triplet-ii/?envType=daily-question&envId=2025-04-03

func maximumTripletValueII(nums []int) int64 {
	n := len(nums)
	res, imax, dmax := 0, 0, 0
	for k := 0; k < n; k++ {
		res = max(res, dmax*nums[k])
		dmax = max(dmax, imax-nums[k])
		imax = max(imax, nums[k])
	}
	return int64(res)
}
