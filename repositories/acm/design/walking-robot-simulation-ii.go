package design

type Robot struct {
	x      int
	y      int
	d      [][]int
	width  int
	height int
	idx    int
}

func NewRobot(width int, height int) Robot {
	d := [][]int{
		{1, 0},
		{0, 1},
		{-1, 0},
		{0, -1},
	}
	return Robot{
		width:  width,
		height: height,
		x:      0,
		y:      0,
		idx:    0,
		d:      d,
	}
}

func (this *Robot) Step(num int) {
	idx := this.idx
	for num != 0 {
		dx, dy := this.d[idx][0], this.d[idx][1]
		nx := this.x + dx*num
		ny := this.y + dy*num
		if nx >= this.width {
			num = nx - this.width + 1
			nx = this.width - 1
			this.x = nx
			this.y = ny
			idx = (idx + 1) % 4
			this.idx = idx
			continue
		}
		if ny >= this.height {
			num = ny - this.height + 1
			ny = this.height - 1
			this.x = nx
			this.y = ny
			idx = (idx + 1) % 4
			this.idx = idx
		}
		if nx < 0 {
			num = -nx
			nx = 0
			this.x = nx
			this.y = ny
			idx = (idx + 1) % 4
			this.idx = idx
			continue
		}
		if ny < 0 {
			num = -ny
			ny = 0
			this.x = nx
			this.y = ny
			idx = (idx + 1) % 4
			this.idx = idx
		}
		this.x = nx
		this.y = ny
		num = 0
	}
}

func (this *Robot) GetPos() []int {
	return []int{this.x, this.y}
}

func (this *Robot) GetDir() string {
	switch this.idx {
	case 0:
		return "East"
	case 1:
		return "North"
	case 2:
		return "West"
	case 3:
		return "South"
	}
	return ""
}

/**
 * Your Robot object will be instantiated and called as such:
 * obj := Constructor(width, height);
 * obj.Step(num);
 * param_2 := obj.GetPos();
 * param_3 := obj.GetDir();
 */
