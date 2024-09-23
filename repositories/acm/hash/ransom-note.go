package hash

func canConstruct(ransomNote string, magazine string) bool {
	m := map[byte]int{}
	for i := 0; i < len(magazine); i++ {
		m[magazine[i]]++
	}
	for i := 0; i < len(ransomNote); i++ {
		m[ransomNote[i]]--
		if m[ransomNote[i]] < 0 {
			return false
		}
	}
	return true
}
