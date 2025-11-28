package math

func countOperations(num1 int, num2 int) int {
	cnt := 0
	for num1 != 0 && num2 != 0 {
		num1, num2 = max(num1, num2), min(num1, num2)
		cnt += num1 / num2
		num1 %= num2
	}
	return cnt
}
