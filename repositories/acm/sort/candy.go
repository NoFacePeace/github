package sort

import "sort"

func candy(ratings []int) int {
	n := len(ratings)
	arr := [][2]int{}
	for k, v := range ratings {
		arr = append(arr, [2]int{k, v})
	}
	sort.Slice(arr, func(a, b int) bool {
		return arr[a][1] < arr[b][1]
	})
	sogas := make([]int, n)
	for _, v := range arr {
		idx, rate := v[0], v[1]
		if idx == 0 {
			if rate == ratings[idx+1] {
				sogas[idx] = 1
			} else {
				sogas[idx] = sogas[idx+1] + 1
			}
			continue
		}
		if idx == n-1 {
			if rate == ratings[idx-1] {
				sogas[idx] = 1
			} else {
				sogas[idx] = sogas[idx-1] + 1
			}
			continue
		}
		if rate == ratings[idx-1] && rate == ratings[idx+1] {
			sogas[idx] = 1
			continue
		}
		if rate == ratings[idx-1] {
			sogas[idx] = sogas[idx+1] + 1
			continue
		}
		if rate == ratings[idx+1] {
			sogas[idx] = sogas[idx-1] + 1
			continue
		}
		soga := max(sogas[idx-1], sogas[idx+1])
		sogas[idx] = soga + 1
	}
	sum := 0
	for _, v := range sogas {
		sum += v
	}
	return sum
}
