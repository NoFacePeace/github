package array

func maxDivScore(nums []int, divisors []int) int {
	n := len(divisors)
	if n == 0 {
		return 0
	}
	arr := make([]int, n)
	for i := 0; i < len(divisors); i++ {
		cnt := 0
		for j := 0; j < len(nums); j++ {
			if nums[j]%divisors[i] == 0 {
				cnt++
			}
		}
		arr[i] = cnt
	}
	value := arr[0]
	idx := 0
	for k, v := range arr {
		if v < value {
			continue
		}
		if v > value {
			value = v
			idx = k
			continue
		}
		if divisors[idx] > divisors[k] {
			value = v
			idx = k
		}
	}
	return divisors[idx]
}
