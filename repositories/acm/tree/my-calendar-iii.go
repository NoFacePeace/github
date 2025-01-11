package tree

type MyCalendarThree map[int]pair

func NewMyCalendarThree() MyCalendarThree {
	return MyCalendarThree{}
}

func (t MyCalendarThree) update(start, end, l, r, idx int) {
	if r < start || end < l {
		return
	}
	if start <= l && r <= end {
		p := t[idx]
		p.first++
		p.second++
		t[idx] = p
		return
	}
	mid := (l + r) / 2
	t.update(start, end, l, mid, idx*2)
	t.update(start, end, mid+1, r, idx*2+1)
	p := t[idx]
	p.first = p.second + max(t[idx*2].first, t[idx*2+1].first)
	t[idx] = p
}

func (t MyCalendarThree) Book(start int, end int) int {
	t.update(start, end-1, 0, 1e9, 1)
	return t[1].first
}
