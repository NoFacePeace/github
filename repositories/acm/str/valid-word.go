package str

func isValid(word string) bool {
	vowel := false
	consonant := false
	for _, v := range word {
		if !((v >= '0' && v <= '9') || (v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z')) {
			return false
		}
		if v >= '0' && v <= '9' {
			continue
		}
		if v == 'a' || v == 'i' || v == 'e' || v == 'u' || v == 'o' || v == 'A' || v == 'I' || v == 'E' || v == 'U' || v == 'O' {
			vowel = true
		} else {
			consonant = true
		}
	}
	return len(word) >= 3 && vowel && consonant
}
