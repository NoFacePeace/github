package math

func replaceNonCoprimes(nums []int) []int {
	var gcd func(int, int) int
	gcd = func(a, b int) int {
		if a < b {
			a, b = b, a
		}
		if a%b == 0 {
			return b
		}
		a = a % b
		return gcd(a, b)
	}
	ans := []int{}
	n := len(nums)
	first := nums[0]
	for i := 1; i < n; i++ {
		num := gcd(first, nums[i])
		if num == 1 {
			for len(ans) > 0 {
				num := gcd(ans[len(ans)-1], first)
				if num == 1 {
					break
				}
				first = first * ans[len(ans)-1] / num
				ans = ans[:len(ans)-1]
			}
			ans = append(ans, first)
			first = nums[i]
			continue
		}
		first = first * nums[i] / num
	}
	for len(ans) > 0 {
		num := gcd(ans[len(ans)-1], first)
		if num == 1 {
			break
		}
		first = first * ans[len(ans)-1] / num
		ans = ans[:len(ans)-1]
	}
	ans = append(ans, first)
	return ans
}
