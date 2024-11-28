package search

func canReachCorner(xCorner int, yCorner int, circles [][]int) bool {
	visited := make([]bool, len(circles))

	var dfs func(i int) bool
	dfs = func(i int) bool {
		x1, y1, r1 := circles[i][0], circles[i][1], circles[i][2]
		if bottom(x1, y1, r1, xCorner, yCorner) {
			return true
		}
		visited[i] = true
		for j := 0; j < len(circles); j++ {
			if !visited[j] && cross(x1, y1, r1, circles[j][0], circles[j][1], circles[j][2], xCorner, yCorner) && dfs(j) {
				return true
			}
		}
		return false
	}
	for i := range circles {
		x, y, r := circles[i][0], circles[i][1], circles[i][2]
		if in(0, 0, x, y, r) || in(xCorner, yCorner, x, y, r) {
			return false
		}
		if !visited[i] && top(x, y, r, xCorner, yCorner) && dfs(i) {
			return false
		}
	}
	return true
}

func in(px, py, x, y, r int) bool {
	return (x-px)*(x-px)+(y-py)*(y-py) <= r*r
}

func top(x, y, r, xc, yc int) bool {
	return (abs(x) <= r && 0 <= y && y <= yc) || (0 <= x && x <= xc && abs(y-yc) <= r) || in(x, y, 0, yc, r)
}

func bottom(x, y, r, xc, yc int) bool {
	return (abs(y) <= r && 0 <= x && x <= xc) || (0 <= y && y <= yc && abs(x-xc) <= r) || in(x, y, xc, 0, r)
}

func cross(x1, y1, r1, x2, y2, r2, xc, yc int) bool {
	return (x1-x2)*(x1-x2)+(y1-y2)*(y1-y2) <= (r1+r2)*(r1+r2) && x1*r2+x2*r1 < (r1+r2)*xc && y1*r2+y2*r1 < (r1+r2)*yc
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
