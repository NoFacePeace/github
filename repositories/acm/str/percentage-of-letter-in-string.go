package str

// https://leetcode.cn/problems/percentage-of-letter-in-string/solutions/1538681/zi-mu-zai-zi-fu-chuan-zhong-de-bai-fen-b-6jm6/?envType=daily-question&envId=2025-03-31

func percentageLetter(s string, letter byte) int {
	m := map[byte]int{}
	n := len(s)
	for i := 0; i < n; i++ {
		m[s[i]]++
	}
	return m[letter] * 100 / n
}
