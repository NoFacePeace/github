package greedy

import "strconv"

func maximum69Number(num int) int {
	str := strconv.Itoa(num)
	arr := []byte(str)
	for i := 0; i < len(arr); i++ {
		if arr[i] == '6' {
			arr[i] = '9'
			break
		}
	}
	str = string(arr)
	num, _ = strconv.Atoi(str)
	return num
}
