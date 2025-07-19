package sort

import (
	"sort"
	"strings"
)

func removeSubfolders(folder []string) []string {
	m := map[string]bool{}
	sort.Slice(folder, func(a, b int) bool {
		return len(folder[a]) < len(folder[b])
	})
	for _, v := range folder {
		status := true
		arr := strings.Split(v, "/")
		n := len(arr)
		str := ""
		for i := 1; i < n; i++ {
			str += "/" + arr[i]
			if _, ok := m[str]; ok {
				m[str] = true
				status = false
				break
			}
		}
		m[v] = status
	}
	ans := []string{}
	for k, v := range m {
		if v {
			ans = append(ans, k)
		}
	}
	return ans
}
