package set

func distinctNames(ideas []string) int64 {
	names := make(map[byte]map[string]struct{})
	for _, idea := range ideas {
		if _, exists := names[idea[0]]; !exists {
			names[idea[0]] = make(map[string]struct{})
		}
		names[idea[0]][idea[1:]] = struct{}{}
	}
	getIntersectSize := func(a, b map[string]struct{}) int {
		count := 0
		for key := range a {
			if _, found := b[key]; found {
				count++
			}
		}
		return count
	}
	var ans int64 = 0
	for preA, setA := range names {
		for preB, setB := range names {
			if preB == preA {
				continue
			}
			insersect := getIntersectSize(setA, setB)
			ans += int64(len(setA)-insersect) * int64(len(setB)-insersect)
		}
	}
	return ans
}
