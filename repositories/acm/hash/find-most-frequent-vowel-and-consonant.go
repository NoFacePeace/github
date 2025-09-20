package hash

func maxFreqSum(s string) int {
	m := map[rune]int{}
	for _, v := range s {
		m[v]++
	}
	vowel := 0
	consonant := 0
	for k, v := range m {
		if k == 'a' || k == 'e' || k == 'i' || k == 'o' || k == 'u' {
			vowel = max(vowel, v)
			continue
		}
		consonant = max(consonant, v)
	}
	return vowel + consonant
}
