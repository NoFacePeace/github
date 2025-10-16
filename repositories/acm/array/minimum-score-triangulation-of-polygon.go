package array

import "math"

func minScoreTriangulation(values []int) int {
	var f func(arr []int) int
	f = func(arr []int) int {
		n := len(arr)
		if n < 3 {
			return 0
		}
		if n == 3 {
			return arr[0] * arr[1] * arr[2]
		}
		ans := math.MaxInt
		for i := 0; i < 3; i++ {
			j := i
			sum := 0
			tmp := []int{}
			for {
				tmp = append(tmp, arr[j%n])
				sum += arr[j%n] * arr[(j+1)%n] * arr[(j+2)%n]
				j += 2
				if j+2 > i+n {
					if j%n != i {
						tmp = append(tmp, arr[j%n])
					}

					break
				}
			}
			sum += f(tmp)
			ans = min(ans, sum)
		}
		return ans
	}
	return f(values)
}
