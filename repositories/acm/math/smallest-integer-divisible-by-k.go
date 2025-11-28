package math

func smallestRepunitDivByK(k int) int {
	cnt := 1
	num := 1
	for num < k {
		num = num*10 + 1
		cnt++
	}
	visited := map[int]bool{}
	for num%k != 0 {
		if visited[num] {
			return -1
		}
		visited[num] = true
		mod := num % k
		mod = mod*10 + 1
		num = mod
		cnt++
	}
	return cnt
}
