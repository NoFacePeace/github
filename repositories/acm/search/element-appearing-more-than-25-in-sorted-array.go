package search

func findSpecialInteger(arr []int) int {
	bs := func(target int) int {
		lo, hi := 0, len(arr)-1
		res := len(arr)
		for lo <= hi {
			mid := (lo + hi) / 2
			if arr[mid] >= target {
				res = mid
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		}
		return res
	}
	n := len(arr)
	span := n/4 + 1
	for i := 0; i < n; i += span {
		start := bs(arr[i])
		end := bs(arr[i] + 1)
		if end-start >= span {
			return arr[i]
		}
	}
	return -1
}
