package set

// https://leetcode.cn/problems/range-addition-ii/description/
func maxCount(m int, n int, ops [][]int) int {
	row := m
	column := n
	for _, op := range ops {
		row = min(op[0], row)
		column = min(op[1], column)
	}
	return row * column
}
