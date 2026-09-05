package array

func firstStableIndex(nums []int, k int) int {
	maxNums := []int{}
	mx := 0
	for k, v := range nums {
		if k == 0 {
			maxNums = append(maxNums, v)
			mx = v
			continue
		}
		mx = max(mx, v)
		maxNums = append(maxNums, mx)
	}
	n := len(nums)
	minNums := make([]int, n)
	mn := 0
	for i := n - 1; i >= 0; i-- {
		if i == n-1 {
			mn = nums[i]
			minNums[i] = mn
			continue
		}
		mn = min(mn, nums[i])
		minNums[i] = mn
	}
	ans := -1
	mn = 0
	for i := 0; i < n; i++ {
		val := maxNums[i] - minNums[i]
		if val > k {
			continue
		}
		if ans == -1 {
			ans = i
			mn = val
			break
		}
	}
	return ans
}
