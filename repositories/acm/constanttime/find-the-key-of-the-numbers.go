package constanttime

func generateKey(num1 int, num2 int, num3 int) int {
	ans := 0
	bit := 1
	for i := 0; i < 4; i++ {
		num := num1 % 10
		num1 /= 10
		num = min(num, num2%10)
		num2 /= 10
		num = min(num, num3%10)
		num3 /= 10
		num *= bit
		bit *= 10
		ans += num
	}
	return ans
}
