package math

func numSub(s string) int {
	arr := []string{}
	sub := ""
	n := len(s)
	for i := 0; i < n; i++ {
		if s[i] == '1' {
			sub += s[i : i+1]
			continue
		}
		if sub == "" {
			continue
		}
		arr = append(arr, sub)
		sub = ""
	}
	if sub != "" {
		arr = append(arr, sub)
	}
	mod := int(1e9) + 7
	ans := 0
	for i := 0; i < len(arr); i++ {
		l := len(arr[i])
		cnt := l * (l + 1) / 2
		ans += cnt
		ans %= mod
	}
	return ans
}
