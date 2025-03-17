package search

// https://leetcode.cn/problems/range-frequency-queries/description/

type RangeFreqQuery struct {
	occurrence map[int][]int
}

func Constructor(arr []int) RangeFreqQuery {
	occurrence := make(map[int][]int)
	for i, v := range arr {
		occurrence[v] = append(occurrence[v], i)
	}
	return RangeFreqQuery{occurrence: occurrence}
}

func (this *RangeFreqQuery) Query(left int, right int, value int) int {
	pos, exists := this.occurrence[value]
	if !exists {
		return 0
	}
	l := lowerBound(pos, left)
	r := upperBound(pos, right)
	return r - l
}

func lowerBound(pos []int, target int) int {
	low, high := 0, len(pos)-1
	for low <= high {
		mid := low + (high-low)/2
		if pos[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low
}

func upperBound(pos []int, target int) int {
	low, high := 0, len(pos)-1
	for low <= high {
		mid := low + (high-low)/2
		if pos[mid] <= target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low
}

/**
 * Your RangeFreqQuery object will be instantiated and called as such:
 * obj := Constructor(arr);
 * param_1 := obj.Query(left,right,value);
 */
