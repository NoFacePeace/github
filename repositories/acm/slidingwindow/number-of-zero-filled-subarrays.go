package slidingwindow

func zeroFilledSubarray(nums []int) int64 {
	arr := []int{}
	cnt := 0
	for _, v := range nums {
		if v == 0 {
			cnt++
			continue
		}
		if cnt != 0 {
			arr = append(arr, cnt)
		}
		cnt = 0
	}
	if cnt != 0 {
		arr = append(arr, cnt)
	}
	ans := 0
	n := len(arr)
	for i := 0; i < n; i++ {
		num := arr[i]
		ans += num * (num + 1) / 2
	}
	return int64(ans)
}
