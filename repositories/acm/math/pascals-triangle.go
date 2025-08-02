package math

func generate(numRows int) [][]int {
	ans := [][]int{}
	for i := 0; i < numRows; i++ {
		if i == 0 {
			ans = append(ans, []int{1})
			continue
		}
		tmp := []int{}
		for j := 0; j <= len(ans[i-1]); j++ {
			if j == 0 {
				tmp = append(tmp, 1)
				continue
			}
			if j == len(ans[i-1]) {
				tmp = append(tmp, 1)
				continue
			}
			tmp = append(tmp, ans[i-1][j-1]+ans[i-1][j])
		}
		ans = append(ans, tmp)
	}
	return ans
}
