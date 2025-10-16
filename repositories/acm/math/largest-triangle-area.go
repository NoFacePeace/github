package math

import (
	"fmt"
	"math"
)

func largestTriangleArea(points [][]int) float64 {
	n := len(points)
	dist := func(i, j int) float64 {
		x1, y1 := points[i][0], points[i][1]
		x2, y2 := points[j][0], points[j][1]
		if x1 < x2 {
			x1, x2 = x2, x1
		}
		if y1 < y2 {
			y1, y2 = y2, y1
		}
		return math.Sqrt(float64((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)))
	}
	compute := func(i, j, k int) float64 {
		if i == 18 && j == 20 && k == 21 {
			fmt.Println()
		}
		d1 := dist(i, j)
		d2 := dist(i, k)
		d3 := dist(j, k)
		d4 := (d1 + d2 + d3) / 2
		d4 = max(d4, d1)
		d4 = max(d4, d2)
		d4 = max(d4, d3)
		return math.Sqrt(d4 * (d4 - d1) * (d4 - d2) * (d4 - d3))
	}
	ans := 0.0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			for k := j + 1; k < n; k++ {
				area := compute(i, j, k)
				ans = max(ans, area)
				if math.IsNaN(ans) {
					fmt.Println(ans)
				}
			}
		}
	}
	return ans
}
