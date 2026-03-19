package array

func minSwaps(grid [][]int) int {
	arr := []int{}
	n := len(grid)
	for i := 0; i < n; i++ {
		cnt := 0
		for j := n - 1; j >= 0; j-- {
			if grid[i][j] == 1 {
				break
			}
			cnt++
		}
		arr = append(arr, cnt)
	}
	ans := 0
	for i := 0; i < n; i++ {
		if arr[i] >= n-i-1 {
			continue
		}
		j := i + 1
		for j < n {
			if arr[j] >= n-i-1 {
				break
			}
			arr[i], arr[j] = arr[j], arr[i]
			j++
		}
		if j == n {
			return -1
		}
		arr[i], arr[j] = arr[j], arr[i]
		ans += j - i
	}
	return ans
}
