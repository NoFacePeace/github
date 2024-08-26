package array

func minimumValueSum(nums []int, andValues []int) int {
	const INF = (1 << 20) - 1
	n := len(nums)
	m := len(andValues)
	memo := make([]map[int]int, n*m)
	for i := range memo {
		memo[i] = make(map[int]int)
	}

	var dfs func(i, j, cur int) int
	dfs = func(i, j, cur int) int {
		key := i*m + j
		if i == n && j == m {
			return 0
		}
		if i == n || j == m {
			return INF
		}
		if val, ok := memo[key][cur]; ok {
			return val
		}

		cur &= nums[i]
		if cur&andValues[j] < andValues[j] {
			return INF
		}

		res := dfs(i+1, j, cur)
		if cur == andValues[j] {
			res = min(res, dfs(i+1, j+1, INF)+nums[i])
		}
		memo[key][cur] = res
		return res
	}

	res := dfs(0, 0, INF)
	if res < INF {
		return res
	}
	return -1
}
