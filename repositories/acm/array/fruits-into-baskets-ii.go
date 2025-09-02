package array

func numOfUnplacedFruits(fruits []int, baskets []int) int {
	for i := 0; i < len(fruits); i++ {
		tmp := []int{}
		j := 0
		for ; j < len(baskets); j++ {
			if fruits[i] <= baskets[j] {
				break
			}
			tmp = append(tmp, baskets[j])
		}
		if j < len(baskets) {
			tmp = append(tmp, baskets[j+1:]...)
		}
		baskets = tmp
	}
	return len(baskets)
}
