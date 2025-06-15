package math

func minMaxDifference(num int) int {
	arr := []int{}
	for num > 0 {
		bit := num % 10
		arr = append(arr, bit)
		num /= 10
	}
	n := len(arr)
	if n == 0 {
		return 0
	}
	a := arr[n-1]
	b := arr[n-1]
	for i := n - 1; i >= 0; i-- {
		if arr[i] != 9 {
			b = arr[i]
			break
		}
	}
	mx := 0
	mn := 0
	tmp := 1
	for _, v := range arr {
		if v == b {
			mx += tmp * 9
		} else {
			mx += tmp * v
		}
		if v == a {
			mn += 0 * v
		} else {
			mn += tmp * v
		}
		tmp *= 10
	}
	return mx - mn
}
