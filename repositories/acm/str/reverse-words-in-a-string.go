package str

import "strings"

func reverseWords(s string) string {
	arr := []string{}
	n := len(s)
	str := ""
	for i := 0; i < n; i++ {
		if s[i] == ' ' {
			if str != "" {
				arr = append(arr, str)
				str = ""
			}
			continue
		}
		str += s[i : i+1]
	}
	if str != "" {
		arr = append(arr, str)
	}
	n = len(arr)
	for i := 0; i < n/2; i++ {
		arr[i], arr[n-i-1] = arr[n-i-1], arr[i]
	}
	return strings.Join(arr, " ")
}
