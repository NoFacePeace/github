package hash

import "sort"

func findXSum(nums []int, k int, x int) []int {
	m := map[int]*findXSumEle{}
	arr := []*findXSumEle{}
	n := len(nums)
	ans := []int{}
	for i := 0; i < n; i++ {
		num := nums[i]
		if _, ok := m[num]; !ok {
			e := &findXSumEle{}
			e.val = num
			e.cnt = 1
			arr = append(arr, e)
			m[num] = e
		} else {
			e := m[num]
			e.cnt++
		}
		if i < k-1 {
			continue
		}
		if i >= k {
			num := nums[i-k]
			e := m[num]
			e.cnt--
		}
		sort.Slice(arr, func(a, b int) bool {
			if arr[a].cnt > arr[b].cnt {
				return true
			}
			if arr[a].cnt < arr[b].cnt {
				return false
			}
			return arr[a].val > arr[b].val
		})
		sum := 0
		cnt := 0
		for i := 0; i < len(arr); i++ {
			if cnt >= x {
				break
			}
			if arr[i].cnt == 0 {
				continue
			}
			cnt++
			sum += arr[i].val * arr[i].cnt
		}
		ans = append(ans, sum)
	}
	return ans
}

type findXSumEle struct {
	val int
	cnt int
}
