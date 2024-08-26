package linkedlist

import "golang.org/x/exp/rand"

type RandomizedSet struct {
	nums    []int
	indices map[int]int
}

func NewRandomizedSet() RandomizedSet {
	return RandomizedSet{[]int{}, map[int]int{}}
}

func (this *RandomizedSet) Insert(val int) bool {
	if _, ok := this.indices[val]; ok {
		return false
	}
	this.indices[val] = len(this.nums)
	this.nums = append(this.nums, val)
	return true
}

func (this *RandomizedSet) Remove(val int) bool {
	id, ok := this.indices[val]
	if !ok {
		return false
	}
	last := len(this.nums) - 1
	this.nums[id] = this.nums[last]
	this.indices[this.nums[id]] = id
	this.nums = this.nums[:last]
	delete(this.indices, val)
	return true
}

func (this *RandomizedSet) GetRandom() int {
	return this.nums[rand.Intn(len(this.nums))]
}

/**
 * Your RandomizedSet object will be instantiated and called as such:
 * obj := Constructor();
 * param_1 := obj.Insert(val);
 * param_2 := obj.Remove(val);
 * param_3 := obj.GetRandom();
 */
