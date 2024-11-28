package array

func numberOfAlternatingGroupsII(colors []int, k int) int {
	n := len(colors)
	if n < k {
		return 0
	}
	cnt := 0
	ans := 0
	colors = append(colors, colors...)
	n += k - 1
	for i := 0; i < n; i++ {
		if i == 0 {
			if colors[i] != colors[i+1] {
				cnt++
			}
			continue
		}
		if i < k-1 {
			if colors[(i-1+n)%n] != colors[i] && colors[i] != colors[(i+1)%n] {
				cnt++
			}
			continue
		}
		if i == k-1 {
			if colors[i] != colors[i-1] && cnt == k-1 {
				ans++
			}
			if colors[(i-1+n)%n] != colors[i] && colors[i] != colors[(i+1)%n] {
				cnt++
			}
			continue
		}
		if colors[i-k] != colors[i-k+1] {
			cnt--
		}
		if colors[i-k+1] != colors[i-k] && colors[i-k+1] != colors[i-k+2] {
			cnt--
		}
		if colors[i-k+1] != colors[i-k+2] {
			cnt++
		}
		if colors[i] != colors[i-1] && cnt == k-1 {
			ans++
		}
		if colors[(i-1+n)%n] != colors[i] && colors[i] != colors[(i+1)%n] {
			cnt++
		}
	}
	return ans
}
