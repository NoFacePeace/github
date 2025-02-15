package greedy

import "sort"

//https://leetcode.cn/problems/maximum-number-of-coins-you-can-get/description/

func maxCoins(piles []int) int {
	sort.Ints(piles)
	n := len(piles)
	ans := 0
	for i := n / 3; i < n; i += 2 {
		ans += piles[i]
	}
	return ans
}
