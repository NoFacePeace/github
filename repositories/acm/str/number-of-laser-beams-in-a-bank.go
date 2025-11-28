package str

func numberOfBeams(bank []string) int {
	arr := []int{}
	for _, str := range bank {
		n := len(str)
		cnt := 0
		for i := 0; i < n; i++ {
			if str[i] == '1' {
				cnt++
			}
		}
		if cnt != 0 {
			arr = append(arr, cnt)
		}
	}
	ans := 0
	n := len(arr)
	for i := 0; i < n-1; i++ {
		ans += arr[i] * arr[i+1]
	}
	return ans
}
