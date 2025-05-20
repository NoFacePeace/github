package greedy

// https://leetcode.cn/problems/longest-unequal-adjacent-groups-subsequence-i/?envType=daily-question&envId=2025-05-15

func getLongestSubsequence(words []string, groups []int) []string {
	ans := []string{}
	n := len(groups)
	if n == 0 {
		return ans
	}
	val := groups[0]
	arr := []int{}
	arr = append(arr, 0)
	for i := 1; i < n; i++ {
		if groups[i] == val {
			continue
		}
		arr = append(arr, i)
		val = groups[i]
	}
	for _, v := range arr {
		ans = append(ans, words[v])
	}
	return ans
}
