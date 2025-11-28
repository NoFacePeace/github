package array

func maxPower(stations []int, r int, k int) int64 {
	n := len(stations)
	cnt := make([]int, n+1)
	for i := 0; i < n; i++ {
		left := max(0, i-r)
		right := min(n, i+r+1)
		cnt[left] += stations[i]
		cnt[right] -= stations[i]
	}
	mn := stations[0]
	total := 0
	for _, s := range stations {
		mn = min(mn, s)
		total += s
	}
	lo, hi := mn, total+k
	res := 0
	check := func(val int) bool {
		n := len(cnt) - 1
		diff := make([]int, len(cnt))
		copy(diff, cnt)
		sum := 0
		remaining := k
		for i := 0; i < n; i++ {
			sum += diff[i]
			if sum < val {
				add := val - sum
				if remaining < add {
					return false
				}
				remaining -= add
				end := min(n, i+2*r+1)
				diff[end] -= add
				sum += add
			}
		}
		return true
	}
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if check(mid) {
			res = mid
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return int64(res)
}
