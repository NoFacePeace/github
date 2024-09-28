package slidingwindow

func takeCharacters(s string, k int) int {
	n := len(s)
	arr := []byte(s)
	left := n - 1
	a, b, c := 0, 0, 0
	cnt := 0
	f := func(byt byte, num int) {
		switch byt {
		case 'a':
			a += num
		case 'b':
			b += num
		case 'c':
			c += num
		}
	}
	for left >= 0 {
		f(arr[left], 1)
		cnt++
		if a >= k && b >= k && c >= k {
			break
		}
		left--
	}
	if left == -1 {
		return -1
	}
	if left == 0 {
		return cnt
	}
	right := n
	arr = append(arr, arr...)
	ans := cnt
	for left < n && right < 2*n {
		if a >= k && b >= k && c >= k {
			ans = min(ans, cnt)
			f(arr[left], -1)
			left++
			cnt--
			continue
		}
		f(arr[right], 1)
		right++
		cnt++
	}
	return ans
}
