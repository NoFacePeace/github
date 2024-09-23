package str

import "strings"

func fullJustify(words []string, maxWidth int) []string {
	return nil
}

func split(words []string, maxWidth int) [][]string {
	arr := [][]string{}
	sub := []string{}
	cnt := 0
	for _, v := range words {
		if cnt+len(v) == maxWidth {
			sub = append(sub, v)
			arr = append(arr, sub)
			sub = nil
			continue
		}
		if cnt+len(v) > maxWidth {
			arr = append(arr, sub)
			sub = append([]string{}, v)
			cnt = len(v) + 1
			continue
		}
		sub = append(sub, v)
		cnt += len(v) + 1
	}
	if len(sub) != 0 {
		arr = append(arr, sub)
	}
	return arr
}

func join(words []string, maxWidth int, end bool) string {
	if end {
		str := strings.Join(words, " ")
		for len(str) < maxWidth {
			str += " "
		}
		return str
	}
	n := len(words)
	if n == 1 {
		str := words[0]
		for len(str) < maxWidth {
			str += " "
		}
		return str
	}
	length := 0
	for i := 0; i < n; i++ {
		length += len(words[i])
	}
	space := (maxWidth - length) / (n - 1)
	extra := (maxWidth - length) % (n - 1)
	str := ""
	for i := 0; i < n; i++ {
		str += words[i]
		if i < n-1 {
			for j := 0; j < space; j++ {
				str += " "
			}
			if extra > 0 {
				str += " "
				extra--
			}
		}
	}
	return str
}
