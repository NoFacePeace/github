package array

func findSmallestInteger(nums []int, value int) int {
	n := len(nums)
	for i := 0; i < n; i++ {
		num := nums[i]
		if num >= 0 {
			continue

		}
		num = -num
		nums[i] += num/value*value + value
	}
	arr := make([][]int, value)
	for i := 0; i < n; i++ {
		num := nums[i]
		mod := num % value
		arr[mod] = append(arr[mod], num)
	}
	for i := 0; i < n; i++ {
		mod := i % value
		if len(arr[mod]) == 0 {
			return i
		}
		arr[mod] = arr[mod][1:]
	}
	return n
}
