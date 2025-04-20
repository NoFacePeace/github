package math

// https://leetcode.cn/problems/count-good-numbers/?envType=daily-question&envId=2025-04-13

func countGoodNumbers(n int64) int {
	mod := int(1e9) + 7
	quick := func(x, y int) int {
		ret := 1
		mul := x
		for y > 0 {
			if y%2 == 1 {
				ret = ret * mul % mod
			}
			mul = mul * mul % mod
			y /= 2
		}
		return ret
	}
	return quick(5, (int(n)+1)/2) * quick(4, int(n)/2) % mod
}
