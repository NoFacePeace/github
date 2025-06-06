package dp

func maxValueOfCoins(piles [][]int, k int) int {
	f := make([]int, k+1)
	for i := range f {
		f[i] = -1
	}
	f[0] = 0
	for _, pile := range piles {
		for i := k; i > 0; i-- {
			value := 0
			for t := 1; t <= len(pile); t++ {
				value += pile[t-1]
				if i >= t && f[i-t] != -1 {
					f[i] = maxSlice(f[i], f[i-t]+value)
				}
			}
		}
	}
	return f[k]
}
