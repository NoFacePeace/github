package str

func minFlips(s string) int {
	n := len(s)
	arr := make([]int, n)
	cnt := 0
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			if s[i] != '1' {
				cnt++
			}
		} else {
			if s[i] != '0' {
				cnt++
			}
		}
		arr[i] = cnt
	}
	ans := cnt
	ans = min(ans, n-cnt)
	for i := 1; i < n; i++ {
		cnt := 0
		if i%2 == 0 {
			cnt = arr[n-1] - arr[i-1]
			if n%2 == 0 {
				cnt += arr[i-1]
			} else {
				cnt += i - arr[i-1]
			}
		} else {
			cnt = arr[n-1] - arr[i-1]
			cnt = n - i - cnt
			if n%2 == 0 {
				cnt += i - arr[i-1]
			} else {
				cnt += arr[i-1]
			}
		}
		ans = min(ans, cnt)
		ans = min(ans, n-cnt)
	}
	return ans
}
