package hash

// https://leetcode.cn/problems/minimum-number-of-operations-to-make-elements-in-array-distinct/solutions/3634685/shi-shu-zu-yuan-su-hu-bu-xiang-tong-suo-cay1s/?envType=daily-question&envId=2025-04-08
func minimumOperations(nums []int) int {
	cnt := 0
	m := map[int]int{}
	for _, v := range nums {
		m[v]++
		if m[v] == 2 {
			cnt++
		}
	}
	if cnt == 0 {
		return 0
	}
	ans := 0
	n := len(nums)
	i := 0
	ok := false
	for ; i < n-3; i += 3 {
		ans++
		m[nums[i]]--
		if m[nums[i]] == 1 {
			cnt--
		}
		m[nums[i+1]]--
		if m[nums[i+1]] == 1 {
			cnt--
		}
		m[nums[i+2]]--
		if m[nums[i+2]] == 1 {
			cnt--
		}
		if cnt == 0 {
			ok = true
			break
		}
	}
	if !ok {
		ans++
	}
	return ans
}
