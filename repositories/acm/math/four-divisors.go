package math

func sumFourDivisors(nums []int) int {
	ans := 0
	for _, num := range nums {
		cnt := 0
		sum := 0
		for i := 1; i*i <= num; i++ {
			if num%i != 0 {
				continue
			}
			if i*i == num {
				cnt++
				sum += i
			} else {
				cnt += 2
				sum += i + num/i
			}
		}
		if cnt == 4 {
			ans += sum
		}
	}
	return ans
}
