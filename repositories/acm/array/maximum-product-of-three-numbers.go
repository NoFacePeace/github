package array

import "math"

func maximumProduct(nums []int) int {
	one := math.MinInt
	two := one
	three := two
	four := math.MaxInt
	five := math.MaxInt
	for _, num := range nums {
		if num < five {
			four = five
			five = num
		} else if num < four {
			four = num
		}
		if num > one {
			three = two
			two = one
			one = num
			continue
		}
		if num > two {
			three = two
			two = num
			continue
		}
		if num > three {
			three = num
			continue
		}

	}
	if four == math.MaxInt {
		return one * two * three
	}
	if five == math.MaxInt {
		return one * two * three
	}
	return max(one*two*three, one*four*five)
}
