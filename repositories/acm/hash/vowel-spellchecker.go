package hash

func spellchecker(wordlist []string, queries []string) []string {
	m := map[string]bool{}
	m1 := map[string]int{}
	m2 := map[string]int{}
	convert := func(s string) string {
		arr := []byte(s)
		for i := 0; i < len(arr); i++ {
			if arr[i] >= 'A' && arr[i] <= 'Z' {
				arr[i] = arr[i] - 'A' + 'a'
			}
		}
		return string(arr)
	}
	convert1 := func(s string) string {
		arr := []byte(s)
		for i := 0; i < len(arr); i++ {
			if arr[i] == 'a' || arr[i] == 'e' || arr[i] == 'i' || arr[i] == 'o' || arr[i] == 'u' {
				arr[i] = '.'
			}
		}
		return string(arr)
	}
	for i, word := range wordlist {
		m[word] = true
		s := convert(word)
		if _, ok := m1[s]; !ok {
			m1[s] = i
		}
		s = convert1(s)
		if _, ok := m2[s]; !ok {
			m2[s] = i
		}
	}
	ans := []string{}
	for _, query := range queries {
		if m[query] {
			ans = append(ans, query)
			continue
		}
		s := convert(query)
		if _, ok := m1[s]; ok {
			ans = append(ans, wordlist[m1[s]])
			continue
		}
		s = convert1(s)
		if _, ok := m2[s]; ok {
			ans = append(ans, wordlist[m2[s]])
			continue
		}
		ans = append(ans, "")
	}
	return ans
}
