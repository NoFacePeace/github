package array

func minTime(skill []int, mana []int) int64 {
	n := len(skill)
	m := len(mana)
	times := make([]int, n)
	for i := 0; i < m; i++ {
		t := 0
		for j := 0; j < n; j++ {
			if i == 0 && j == 0 {
				times[j] = skill[j] * mana[i]
				continue
			}
			if i == 0 {
				times[j] = times[j-1] + skill[j]*mana[i]
				t = times[j]
				continue
			}
			t = max(t, times[j]) + skill[j]*mana[i]
		}
		for j := n - 1; j >= 0; j-- {
			if j == n-1 {
				times[j] = t
				continue
			}
			times[j] = times[j+1] - skill[j+1]*mana[i]
		}
	}
	return int64(times[n-1])
}
