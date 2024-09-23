package array

func countQuadruplets(nums []int) int64 {
	n := len(nums)
	pre := make([]int, n+1)
	ans := 0
	for j := 0; j < n; j++ {
		suf := 0
		for k := n - 1; k > j; k-- {
			if nums[j] > nums[k] {
				ans += pre[nums[k]] * suf
			} else {
				suf++
			}
		}
		for x := nums[j] + 1; x <= n; x++ {
			pre[x]++
		}
	}
	return int64(ans)
}
