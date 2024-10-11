package heap

type heap []int

func (h *heap) Push(num int) {
	*h = append(*h, num)
	h.up()
}

func (h *heap) Pop() int {
	arr := *h
	n := len(arr)
	num := arr[0]
	arr[0] = arr[n-1]
	*h = arr[:n-1]
	h.down()
	return num
}

func (h *heap) up() {
	arr := *h
	n := len(arr)
	i := n - 1
	for i > 0 {
		if arr[i] <= arr[(i-1)/2] {
			break
		}
		arr[(i-1)/2], arr[i] = arr[i], arr[(i-1)/2]
		i = (i - 1) / 2
	}
}

func (h *heap) down() {
	arr := *h
	i := 0
	n := len(arr)
	for i*2+1 < n {
		mx := i*2 + 1
		if i*2+2 < n && arr[mx] < arr[i*2+2] {
			mx = i*2 + 2
		}
		if arr[i] > arr[mx] {
			break
		}
		arr[mx], arr[i] = arr[i], arr[mx]
		i = mx
	}
}
