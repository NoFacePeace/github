package array

func lengthOfLIS(nums []int) int {
	arr := []int{}
	search := func(num int) int {
		l, r := 0, len(arr)-1
		for l <= r {
			mid := (l + r) / 2
			if arr[mid] < num {
				l = mid + 1
				continue
			}
			r = mid - 1
		}
		return l
	}
	for _, v := range nums {
		idx := search(v)
		if idx == len(arr) {
			arr = append(arr, v)
			continue
		}
		arr[idx] = v
	}
	return len(arr)
}
