package pointer

func findWinningPlayer(skills []int, k int) int {
	n := len(skills)
	cnt := 0
	i, lastI := 0, 0
	for i < n {
		j := i + 1
		for j < n && skills[j] < skills[i] && cnt < k {
			j++
			cnt++
		}
		if cnt == k {
			return i
		}
		cnt = 1
		lastI = i
		i = j
	}
	return lastI
}
