package str

func robotWithString(s string) string {
	arr := []byte(s)
	n := len(s)
	q := make([]byte, n)
	for i := n - 1; i >= 0; i-- {
		if i == n-1 {
			q[i] = arr[i]
			continue
		}
		q[i] = min(arr[i], q[i+1])
	}
	tmp := []byte{}
	ans := []byte{}
	for i := 0; i < n; i++ {
		for len(tmp) > 0 {
			if tmp[len(tmp)-1] > q[i] {
				break
			}
			ans = append(ans, tmp[len(tmp)-1])
			tmp = tmp[:len(tmp)-1]
		}
		tmp = append(tmp, arr[i])
	}
	tn := len(tmp)
	for i := tn - 1; i >= 0; i-- {
		ans = append(ans, tmp[i])
	}
	return string(ans)
}
