package heap

import "github.com/NoFacePeace/github/repositories/acm/datastruct/heap"

func clearStars(s string) string {
	arr := []byte(s)
	type item struct {
		index int
		value byte
	}
	h := heap.New(func(a, b item) bool {
		if a.value == b.value {
			return a.index > b.index
		}
		return a.value < b.value
	})
	m := map[int]bool{}
	for k, v := range arr {
		if v != '*' {
			item := item{
				index: k,
				value: v,
			}
			h.PushItem(item)
			continue
		}
		if h.Len() == 0 {
			continue
		}
		item := h.PopItem()
		m[item.index] = true
	}
	ans := []byte{}
	for k, v := range arr {
		if v == '*' {
			continue
		}
		if m[k] {
			continue
		}
		ans = append(ans, v)
	}
	return string(ans)
}
