package constanttime

// https://leetcode.cn/problems/verify-preorder-serialization-of-a-binary-tree/
func isValidSerialization(preorder string) bool {
	n := len(preorder)
	if n == 0 {
		return true
	}
	cnt := 1
	for i := 0; i < n; i++ {
		if preorder[i] == ',' && preorder[i-1] != '#' {
			cnt++
			continue
		}
		if preorder[i] == '#' {
			cnt--
		}
		if cnt <= 0 && i != n-1 {
			return false
		}
	}
	if preorder[n-1] != '#' {
		return false
	}
	if cnt == 0 {
		return true
	}
	return false
}
