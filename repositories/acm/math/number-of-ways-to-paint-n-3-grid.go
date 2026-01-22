package math

func numOfWays(n int) int {
	mod := int(1e9) + 7
	fi0, fi1 := 6, 6
	for i := 2; i <= n; i++ {
		nfi0 := (2*fi0 + 2*fi1) % mod
		nfi1 := (2*fi0 + 3*fi1) % mod
		fi0, fi1 = nfi0, nfi1
	}
	return (fi0 + fi1) % mod
}
