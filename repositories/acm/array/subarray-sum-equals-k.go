package array

func subarraySum(nums []int, k int) int {
	ans := 0
	l, r := 0, 0
	n := len(nums)
	sum := 0
	for r < n {
		sum += nums[r]
		if sum == k {
			ans++
			r++
			continue
		}
		if sum < k {
			r++
			continue
		}
		sum -= nums[l]
		l++
		if l > r {
			r++
		}
	}
	return ans
}
