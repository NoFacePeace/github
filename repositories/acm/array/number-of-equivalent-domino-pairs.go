package array

// https://leetcode.cn/problems/number-of-equivalent-domino-pairs/?envType=daily-question&envId=2025-05-04

func numEquivDominoPairs(dominoes [][]int) int {
	size := 10
	arr := make([][]int, size)
	for i := 0; i < size; i++ {
		arr[i] = make([]int, size)
	}
	ans := 0
	for _, dominoe := range dominoes {
		v1, v2 := dominoe[0], dominoe[1]
		if v1 > v2 {
			v1, v2 = v2, v1
		}
		ans += arr[v1][v2]
		arr[v1][v2]++
	}
	return ans
}
