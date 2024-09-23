package dp

func climbStairs(n int) int {
	one := 1
	two := 1
	for i := 2; i <= n; i++ {
		one, two = two, one+two
	}
	return two
}
