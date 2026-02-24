package bitwise

import "strconv"

func readBinaryWatch(turnedOn int) []string {
	arr := []int{1, 2, 4, 8, 1, 2, 4, 8, 16, 32}
	ans := []string{}
	var dfs func(idx int, nums []int)
	dfs = func(idx int, nums []int) {
		if len(nums) == turnedOn {
			hour := 0
			minute := 0
			for _, v := range nums {
				if v < 4 {
					hour += arr[v]
				} else {
					minute += arr[v]
				}
			}
			if hour > 11 {
				return
			}
			if minute > 59 {
				return
			}
			str := ""
			str += strconv.Itoa(hour) + ":"
			if minute < 10 {
				str += "0" + strconv.Itoa(minute)
			} else {
				str += strconv.Itoa(minute)
			}
			ans = append(ans, str)
			return
		}
		if idx == len(arr) {
			return
		}
		dfs(idx+1, nums)
		nums = append(nums, idx)
		dfs(idx+1, nums)
	}
	dfs(0, []int{})
	return ans
}
