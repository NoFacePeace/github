package array

// https://leetcode.cn/problems/design-browser-history/description/

type BrowserHistory struct {
	pos int
	arr []string
}

func NewBrowserHistory(homepage string) BrowserHistory {
	return BrowserHistory{
		pos: 0,
		arr: []string{homepage},
	}
}

func (this *BrowserHistory) Visit(url string) {
	this.arr = this.arr[:this.pos+1]
	this.arr = append(this.arr, url)
	this.pos++
}

func (this *BrowserHistory) Back(steps int) string {
	if this.pos < steps {
		this.pos = 0
	} else {
		this.pos -= steps
	}
	return this.arr[this.pos]
}

func (this *BrowserHistory) Forward(steps int) string {
	if this.pos+steps < len(this.arr) {
		this.pos += steps
	} else {
		this.pos = len(this.arr) - 1
	}
	return this.arr[this.pos]
}

/**
 * Your BrowserHistory object will be instantiated and called as such:
 * obj := Constructor(homepage);
 * obj.Visit(url);
 * param_2 := obj.Back(steps);
 * param_3 := obj.Forward(steps);
 */
