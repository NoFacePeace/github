package dp

func countPartitions(nums []int, k int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	dp := make([]int, n)
	dp[0] = 1
	mn := []int{nums[0]}
	mx := []int{nums[0]}
	l, r := 0, 1
	cnt := 1
	for ; r < n; r++ {
		num := nums[r]
		for len(mx) > 0 {
			n := len(mx)
			if mx[n-1] >= num {
				break
			}
			mx = mx[:n-1]
		}
		mx = append(mx, num)
		for len(mn) > 0 {
			n := len(mn)
			if mn[n-1] <= num {
				break
			}
			mn = mn[:n-1]
		}
		mn = append(mn, num)
		if mx[0]-mn[0] <= k {
			if l == 0 {
				dp[r] = cnt + 1
			} else {
				dp[r] = cnt + dp[l-1]
			}
			cnt += dp[r]
			continue
		}
		for mx[0]-mn[0] > k {
			cnt -= dp[l]
			if nums[l] == mn[0] {
				mn = mn[1:]
			}
			if nums[l] == mx[0] {
				mx = mx[1:]
			}
			l++
		}
		dp[r] = cnt + dp[l-1]
		cnt += dp[r]
	}
	return dp[n-1]
}
