package hash

type NeighborSum struct {
	grid [][]int
	m    map[int][]int
}

func NewNeighborSum(grid [][]int) NeighborSum {
	m := map[int][]int{}
	n := len(grid)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			val := grid[i][j]
			m[val] = []int{i, j}
		}
	}
	return NeighborSum{
		grid: grid,
		m:    m,
	}
}

func (this *NeighborSum) AdjacentSum(value int) int {
	x, y := this.m[value][0], this.m[value][1]
	sum := 0
	if x != 0 {
		sum += this.grid[x-1][y]
	}
	if y != 0 {
		sum += this.grid[x][y-1]
	}
	if y != len(this.grid)-1 {
		sum += this.grid[x][y+1]
	}
	if x != len(this.grid)-1 {
		sum += this.grid[x+1][y]
	}
	return sum
}

func (this *NeighborSum) DiagonalSum(value int) int {
	x, y := this.m[value][0], this.m[value][1]
	sum := 0
	if x != 0 && y != 0 {
		sum += this.grid[x-1][y-1]
	}
	if y != 0 && x != len(this.grid)-1 {
		sum += this.grid[x+1][y-1]
	}
	if x != 0 && y != len(this.grid)-1 {
		sum += this.grid[x-1][y+1]
	}
	if x != len(this.grid)-1 && y != len(this.grid)-1 {
		sum += this.grid[x+1][y+1]
	}
	return sum
}

/**
 * Your NeighborSum object will be instantiated and called as such:
 * obj := Constructor(grid);
 * param_1 := obj.AdjacentSum(value);
 * param_2 := obj.DiagonalSum(value);
 */
