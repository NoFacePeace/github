package str

func maxConsecutiveAnswers(answerKey string, k int) int {
	queue := []int{}
	arr := []byte(answerKey)
	n := len(arr)
	cnt := 0
	mx := 0
	used := 0
	for i := 0; i < n; i++ {
		if arr[i] == 'T' {
			cnt++
		} else {
			if used < k {
				cnt++
				used++
				queue = append(queue, i)
			} else {
				idx := queue[0]
				queue = queue[1:]
				queue = append(queue, i)
				cnt = i - idx
			}
		}
		if cnt > mx {
			mx = cnt
		}
	}
	cnt = 0
	used = 0
	queue = []int{}
	for i := 0; i < n; i++ {
		if arr[i] == 'F' {
			cnt++
		} else {
			if used < k {
				cnt++
				used++
				queue = append(queue, i)
			} else {
				idx := queue[0]
				queue = queue[1:]
				queue = append(queue, i)
				cnt = i - idx
			}
		}
		if cnt > mx {
			mx = cnt
		}
	}
	return mx
}
