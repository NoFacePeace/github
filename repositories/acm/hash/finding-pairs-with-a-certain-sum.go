package hash

type FindSumPairs struct {
	nums2 []int
	m1    map[int]int
	m2    map[int]int
}

func NewFindSumPairs(nums1 []int, nums2 []int) FindSumPairs {
	m1 := map[int]int{}
	for i := 0; i < len(nums1); i++ {
		num1 := nums1[i]
		m1[num1]++
	}
	m2 := map[int]int{}
	for i := 0; i < len(nums2); i++ {
		num2 := nums2[i]
		m2[num2]++
	}
	return FindSumPairs{
		m1:    m1,
		m2:    m2,
		nums2: nums2,
	}
}

func (this *FindSumPairs) Add(index int, val int) {
	num2 := this.nums2[index]
	this.m2[num2]--
	if this.m2[num2] == 0 {
		delete(this.m2, num2)
	}
	num2 += val
	this.nums2[index] = num2
	this.m2[num2]++
}

func (this *FindSumPairs) Count(tot int) int {
	cnt := 0
	m := this.m1
	for k, v := range m {
		cnt += this.m2[tot-k] * v
	}
	return cnt
}

/**
 * Your FindSumPairs object will be instantiated and called as such:
 * obj := Constructor(nums1, nums2);
 * obj.Add(index,val);
 * param_2 := obj.Count(tot);
 */
