package math

// https://leetcode.cn/problems/divisible-and-non-divisible-sums-difference/submissions/632688474/?envType=daily-question&envId=2025-05-27

func differenceOfSums(n int, m int) int {
	k := n / m
	return n*(n+1)/2 - k*(k+1)*m
}
