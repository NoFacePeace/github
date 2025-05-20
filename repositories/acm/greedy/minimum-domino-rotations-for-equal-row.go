package greedy

// https://leetcode.cn/problems/minimum-domino-rotations-for-equal-row/?envType=daily-question&envId=2025-05-03

func minDominoRotations(tops []int, bottoms []int) int {
	n := len(tops)
	check := func(x int) int {
		a, b := 0, 0
		for i := 0; i < n; i++ {
			if tops[i] != x && bottoms[i] != x {
				return -1
			}
			if tops[i] != x {
				a++
			}
			if bottoms[i] != x {
				b++
			}
		}
		return min(a, b)
	}
	t := check(tops[0])
	p := check(bottoms[0])
	if t == -1 {
		return p
	}
	if p == -1 {
		return t
	}
	return min(t, p)
}
