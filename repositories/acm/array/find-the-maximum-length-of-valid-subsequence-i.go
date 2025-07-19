package array

func maximumLength(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	for i := 0; i < n; i++ {
		nums[i] = nums[i] % 2
	}
	cnt := 1
	val := nums[0]
	for i := 1; i < n; i++ {
		if nums[i] != val {
			cnt++
			val = nums[i]
		}
	}
	ans := cnt
	cnt = 0
	for i := 0; i < n; i++ {
		if nums[i] == 0 {
			cnt++
		}
	}
	ans = max(ans, cnt)
	cnt = 0
	for i := 0; i < n; i++ {
		if nums[i] == 1 {
			cnt++
		}
	}
	ans = max(ans, cnt)
	return ans
}
