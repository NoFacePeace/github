package tree

type SeatManager struct {
	arr []int
	n   int
}

func NewSeaManger(n int) SeatManager {
	arr := make([]int, 4*n)
	var f func(i, l, r int)
	f = func(i, l, r int) {
		if l == r {
			arr[i] = l
			return
		}
		mid := (l + r) / 2
		f(2*i, l, mid)
		f(2*i+1, mid+1, r)
		arr[i] = l
	}
	f(1, 1, n)
	return SeatManager{
		arr: arr,
		n:   n,
	}
}

func (this *SeatManager) Reserve() int {
	mn := this.arr[1]
	this.modify(1, 1, this.n, mn, mn, this.n)
	return mn
}

func (this *SeatManager) Unreserve(seatNumber int) {
	this.modify(1, 1, this.n, seatNumber, seatNumber, seatNumber)
}

func (this *SeatManager) modify(i, l, r, l2, r2, val int) {
	if l2 <= l && r2 >= r {
		this.arr[i] = val
		return
	}
	mid := (l + r) / 2
	if l2 <= mid {
		this.modify(2*i, l, mid, l2, r2, val)
	}
	if r2 >= mid+1 {
		this.modify(2*i+1, mid+1, r, l2, r2, val)
	}
	this.arr[i] = min(this.arr[2*i], this.arr[2*i+1])
}
