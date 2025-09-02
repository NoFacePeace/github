package array

import "sort"

func sortMatrix(grid [][]int) [][]int {
	n := len(grid)
	for i := 1; i < n; i++ {
		x := 0
		y := i
		arr := []int{}
		for y < n {
			arr = append(arr, grid[x][y])
			x++
			y++
		}
		x = 0
		y = i
		idx := 0
		sort.Ints(arr)
		for y < n {
			grid[x][y] = arr[idx]
			x++
			y++
			idx++
		}
	}
	for i := 0; i < n; i++ {
		x := i
		y := 0
		arr := []int{}
		for x < n {
			arr = append(arr, grid[x][y])
			x++
			y++
		}
		x = i
		y = 0
		idx := 0
		sort.Slice(arr, func(a, b int) bool {
			return arr[a] > arr[b]
		})
		for x < n {
			grid[x][y] = arr[idx]
			x++
			y++
			idx++
		}
	}
	return grid
}
