package array

func rowAndMaximumOnes(mat [][]int) []int {
	mx := 0
	idx := 0
	for i, m := range mat {
		cnt := 0
		for _, v := range m {
			if v == 1 {
				cnt++
			}
		}
		if cnt > mx {
			mx = cnt
			idx = i
		}
	}
	return []int{idx, mx}
}
