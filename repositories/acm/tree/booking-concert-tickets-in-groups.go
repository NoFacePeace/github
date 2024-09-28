package tree

type BookMyShow struct {
	n    int
	m    int
	size []int
}

func NewBookMyShow(n int, m int) BookMyShow {
	return BookMyShow{
		n:    n,
		m:    m,
		size: make([]int, m),
	}
}

func (this *BookMyShow) Gather(k int, maxRow int) []int {
	for i := 0; i <= min(maxRow, this.n-1); i++ {
		if this.m-this.size[i] >= k {
			pos := this.size[i]
			this.size[i] += k
			return []int{i, pos}
		}
	}
	return []int{}
}

func (this *BookMyShow) Scatter(k int, maxRow int) bool {
	idle := 0
	for i := 0; i <= min(maxRow, this.n-1); i++ {
		idle += this.m - this.size[i]
		if idle < k {
			continue
		}
		idle := 0
		for j := 0; j < i; j++ {
			idle += this.m - this.size[j]
			this.size[j] = this.m
		}
		this.size[i] += k - idle
		return true
	}
	return false
}

/**
 * Your BookMyShow object will be instantiated and called as such:
 * obj := Constructor(n, m);
 * param_1 := obj.Gather(k,maxRow);
 * param_2 := obj.Scatter(k,maxRow);
 */
