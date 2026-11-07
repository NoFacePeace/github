package array

func findKthLargest(nums []int, k int) int {
	var dfs func(nums []int, k int) int
	dfs = func(nums []int, k int) int {
		n := len(nums)
		if n == 1 {
			return nums[0]
		}

		num := nums[n/2]
		l, r := 0, n-1

		for l <= r {
			for l <= r && nums[l] < num {
				l++
			}
			for l <= r && nums[r] > num {
				r--
			}
			if l <= r {
				nums[l], nums[r] = nums[r], nums[l]
				l++
				r--
			}
		}

		target := n - k // 当前数组中，目标的升序下标

		if target <= r {
			// 保留左侧，扣除被排除部分占据的较大排名。
			return dfs(nums[:r+1], k-(n-r-1))
		}
		if target >= l {
			// 保留右侧，第 k 大的 k 不变。
			return dfs(nums[l:], k)
		}
		return num
	}

	return dfs(nums, k)
}
