package bitwise

import (
	"math/bits"
	"sort"
)

func sortByBits(arr []int) []int {
	sort.Slice(arr, func(i, j int) bool {
		ii := bits.OnesCount(uint(arr[i]))
		jj := bits.OnesCount(uint(arr[j]))
		if ii == jj {
			return arr[i] < arr[j]
		}
		return ii < jj
	})
	return arr
}
