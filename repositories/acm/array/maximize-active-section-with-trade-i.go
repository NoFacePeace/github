package array

func maxActiveSectionsAfterTrade(s string) int {
	arr := []int{}
	n := len(s)
	if n == 0 {
		return 0
	}
	last := s[0]
	cnt := 0
	one := 0
	if last == '1' {
		cnt = -1
		one++
	} else {
		cnt = 1
	}
	for i := 1; i < n; i++ {
		c := s[i]
		if c == '1' {
			one++
		}
		if c == last {
			if c == '1' {
				cnt--
			} else {
				cnt++
			}
			continue
		}
		arr = append(arr, cnt)
		cnt = 0
		if c == '1' {
			cnt--
		} else {
			cnt++
		}
		last = c
	}
	arr = append(arr, cnt)
	if len(arr) <= 2 {
		return one
	}
	ans := one

	n = len(arr)
	for i := 0; i < n-2; i++ {
		if arr[i] < 0 {
			continue
		}
		ans = max(ans, arr[i]+arr[i+2]+one)
		i++
	}
	return ans
}
