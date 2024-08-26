package array

func canPartitionKSubsets(nums []int, k int) bool {
	sum := 0
	m := map[int]int{}
	for _, v := range nums {
		sum += v
		m[v]++
	}
	mid := sum / k
	if mid*k != sum {
		return false
	}
	var dfs func(num int) bool
	dfs = func(num int) bool {
		if num == 0 {
			return true
		}
		for i := num; i > 0; i-- {
			if m[i] == 0 {
				continue
			}
			m[i]--
			if dfs(num - i) {
				return true
			}
			m[i]++
		}
		return false
	}
	for i := 0; i < k; i++ {
		if !dfs(mid) {
			return false
		}
	}
	return true
}
