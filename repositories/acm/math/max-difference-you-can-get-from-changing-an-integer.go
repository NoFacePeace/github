package math

func maxDiff(num int) int {
	arr := []int{}
	for num > 0 {
		digit := num % 10
		num /= 10
		arr = append(arr, digit)
	}
	n := len(arr)
	digit := arr[n-1]
	mx := digit
	for i := n - 1; i >= 0; i-- {
		if arr[i] != 9 {
			mx = arr[i]
			break
		}
	}
	mn := digit
	for i := n - 1; i >= 0; i-- {
		if arr[i] != 1 && arr[i] != 0 {
			mn = arr[i]
			break
		}
	}
	a, b := 0, 0
	tmp := 1
	for _, v := range arr {
		if v == mx {
			a += tmp * 9
		} else {
			a += tmp * v
		}
		if v == mn {
			if mn == digit {
				b += tmp * 1
			} else {
				b += tmp * 0
			}

		} else {
			b += tmp * v
		}
		tmp *= 10
	}
	return a - b
}
