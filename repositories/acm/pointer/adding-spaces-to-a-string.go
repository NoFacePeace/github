package pointer

// https://leetcode.cn/problems/adding-spaces-to-a-string/?envType=daily-question&envId=2025-03-30
func addSpaces(s string, spaces []int) string {
	arr := []byte{}
	n := len(s)
	j := 0
	m := len(spaces)
	for i := 0; i < n; i++ {
		if j < m && i == spaces[j] {
			arr = append(arr, ' ')
			j++
		}
		arr = append(arr, s[i])
	}
	return string(arr)
}
