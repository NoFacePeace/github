package array

func imageSmoother(img [][]int) [][]int {
	m := len(img)
	if m == 0 {
		return nil
	}
	n := len(img[0])
	ans := make([][]int, m)
	for i := 0; i < m; i++ {
		ans[i] = make([]int, n)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			sum := 0
			cnt := 0
			for x := -1; x <= 1; x++ {
				if i+x < 0 {
					continue
				}
				if i+x >= m {
					continue
				}
				for y := -1; y <= 1; y++ {
					if j+y < 0 {
						continue
					}
					if j+y >= n {
						continue
					}
					sum += img[i+x][j+y]
					cnt++
				}
			}
			ans[i][j] = sum / cnt
		}
	}
	return ans
}
