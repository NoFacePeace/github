package bitwise

import "math/bits"

func concatenatedBinary(n int) int {
	mod := int(1e9) + 7
	ans := 0
	for i := 1; i <= n; i++ {
		w := bits.Len(uint(i))
		ans = (ans<<w | i) % mod
	}
	return ans
}
