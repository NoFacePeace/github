package str

func countOfSubstrings(word string, k int) int {
	arr := []byte(word)
	n := len(arr)

	ans := 0
	for i := 0; i < n; i++ {
		cnt := 0
		other := 0
		m := map[byte]int{
			'a': 0,
			'e': 0,
			'i': 0,
			'o': 0,
			'u': 0,
		}
		for j := i; j < n; j++ {
			c := arr[j]
			if _, ok := m[c]; ok {
				m[c]++
				if m[c] == 1 {
					cnt++
				}
			} else {
				other++
			}
			if cnt == 5 && other == k {
				ans++
			}
			if other > k {
				break
			}
		}
	}
	return ans
}
