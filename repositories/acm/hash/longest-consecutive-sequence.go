package hash

func longestConsecutive(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	m := map[int]bool{}
	for _, v := range nums {
		m[v] = true
	}
	mx := 0
	num := nums[0]
	for len(m) > 0 {
		cnt := 1
		delete(m, num)
		for i := num + 1; ; i++ {
			if m[i] {
				cnt++
				delete(m, i)
				continue
			}
			break
		}
		for i := num - 1; ; i-- {
			if m[i] {
				cnt++
				delete(m, i)
				continue
			}
			break
		}
		mx = max(mx, cnt)
		for k := range m {
			num = k
			break
		}
	}
	return mx
}
