package hash

import "sort"

func groupAnagrams(strs []string) [][]string {
	m := map[string][]string{}
	for _, v := range strs {
		arr := []byte(v)
		sort.Slice(arr, func(a, b int) bool {
			return arr[a] < arr[b]
		})
		m[string(arr)] = append(m[string(arr)], v)
	}
	ans := [][]string{}
	for _, v := range m {
		ans = append(ans, v)
	}
	return ans
}
