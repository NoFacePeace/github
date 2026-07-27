package array

func maxProduct(nums []int) int {
	one, two := 0, 0
	for _, num := range nums {
		if num > one {
			two = one
			one = num
			continue
		}
		if num > two {
			two = num
		}
	}
	return (one - 1) * (two - 1)
}
