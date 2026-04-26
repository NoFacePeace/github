package array

func closestTarget(words []string, target string, startIndex int) int {
	n := len(words)
	ans := n
	for k, v := range words {
		if v != target {
			continue
		}
		if k <= startIndex {
			ans = min(ans, startIndex-k)
			ans = min(ans, n-startIndex+k)
			continue
		}
		ans = min(ans, k-startIndex)
		ans = min(ans, n-k+startIndex)
	}
	if ans == n {
		return -1
	}
	return ans
}
