package str

// https://leetcode.cn/problems/push-dominoes/?envType=daily-question&envId=2025-05-02

func pushDominoes(dominoes string) string {
	arr := []byte(dominoes)
	i := 0
	n := len(arr)
	l := 0
	r := -1
	for i < n {
		c := arr[i]
		if c == '.' {
			i++
			continue
		}
		if c == 'R' && r == -1 {
			r = i
			i++
			continue
		}
		if c == 'R' {
			for j := r + 1; j < i; j++ {
				arr[j] = 'R'
			}
			r = i
			i++
			continue
		}
		if r == -1 {
			for j := l; j < i; j++ {
				arr[j] = 'L'
			}
			l = i
			i++
			continue
		}
		m := i - r - 1
		for j := 1; j <= m/2; j++ {
			arr[r+j] = 'R'
			arr[i-j] = 'L'
		}
		r = -1
		l = i
		i++
	}
	if r != -1 && l <= r {
		for i := r + 1; i < n; i++ {
			arr[i] = 'R'
		}
	}
	return string(arr)
}
