package bitwise

func minPartitions(n string) int {
	bit := '0'
	for _, v := range n {
		bit = max(bit, v)
	}
	return int(bit - '0')
}
