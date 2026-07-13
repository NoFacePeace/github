package bitwise

func sequentialDigits(low int, high int) []int {
	ans := []int{}
	num := 1
	inc := 1
	for num*10+num+1 < low {
		num = num*10 + num%10 + 1
		inc = inc*10 + 1
	}
	last := num
	for num <= high && last <= high {
		if last >= low {
			ans = append(ans, last)
		}
		last += inc
		if last%10 == 0 {
			num = num*10 + num%10 + 1
			last = num
			inc = inc*10 + 1
		}
	}
	return ans
}
