package math

// https://leetcode.cn/problems/finding-3-digit-even-numbers/solutions/1140756/zhao-chu-3-wei-ou-shu-by-leetcode-soluti-hptf/?envType=daily-question&envId=2025-05-12

func findEvenNumbers(digits []int) []int {
	m := map[int]int{}
	for _, v := range digits {
		m[v]++
	}
	ans := []int{}
	for i := 1; i < 10; i++ {
		if _, ok := m[i]; !ok {
			continue
		}
		for j := 0; j < 10; j++ {
			if _, ok := m[j]; !ok {
				continue
			}
			if i == j && m[j] == 1 {
				continue
			}
			for k := 0; k < 10; k += 2 {
				if _, ok := m[k]; !ok {
					continue
				}
				if k == i && m[k] == 1 {
					continue
				}
				if k == j && m[k] == 1 {
					continue
				}
				if k == i && i == j && m[k] == 2 {
					continue
				}
				ans = append(ans, i*100+j*10+k)
			}
		}
	}
	return ans
}
