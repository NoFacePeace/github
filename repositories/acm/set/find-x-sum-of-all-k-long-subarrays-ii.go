package set

import (
	"github.com/emirpasic/gods/v2/trees/redblacktree"
)

func findXSum(nums []int, k int, x int) []int64 {
	s := NewFindXSumStruct(x)
	ans := []int{}
	n := len(nums)
	for i := 0; i < n; i++ {
		num := nums[i]
		s.Insert(num)
		if i >= k {
			s.Remove(nums[i-k])
		}
		if i >= k-1 {
			ans = append(ans, s.Get())
		}
	}
	return nil
}

type findXSumStruct struct {
	x      int
	result int
	large  *redblacktree.Tree[findXSumPair, struct{}]
	small  *redblacktree.Tree[findXSumPair, struct{}]
	occ    map[int]int
}

func NewFindXSumStruct(x int) *findXSumStruct {
	return &findXSumStruct{
		x:      x,
		result: 0,
		large:  redblacktree.NewWith[findXSumPair, struct{}](findXSumPairComparator),
		small:  redblacktree.NewWith[findXSumPair, struct{}](findXSumPairComparator),
		occ:    map[int]int{},
	}
}

func (f *findXSumStruct) Insert(num int) {
	if f.occ[num] > 0 {
		f.remove(findXSumPair{freq: f.occ[num], num: num})
	}
	f.occ[num]++
	f.insert(findXSumPair{freq: f.occ[num], num: num})
}

func (f *findXSumStruct) Remove(num int) {
	f.remove(findXSumPair{freq: f.occ[num], num: num})
	f.occ[num]--
	if f.occ[num] > 0 {
		f.insert(findXSumPair{freq: f.occ[num], num: num})
	}
}

func (f *findXSumStruct) Get() int {
	return f.result
}

func (f *findXSumStruct) insert(p findXSumPair) {
	if f.large.Size() < f.x {
		f.result += p.freq * p.num
		f.large.Put(p, struct{}{})
		return
	}
	mn := f.large.Left().Key
	if findXSumPairComparator(p, mn) < 0 {
		f.small.Put(p, struct{}{})
		return
	}
	f.result += p.freq * p.num
	f.large.Put(p, struct{}{})
	f.result -= mn.freq * mn.num
	f.large.Remove(mn)
	f.small.Put(mn, struct{}{})
}

func (f *findXSumStruct) remove(p findXSumPair) {
	if _, found := f.large.Get(p); found {
		f.result -= p.freq * p.num
		f.large.Remove(p)
		if f.small.Size() > 0 {
			mx := f.small.Right().Key
			f.result += mx.freq * mx.num
			f.small.Remove(mx)
			f.large.Put(mx, struct{}{})
		}
		return
	}
	if _, found := f.small.Get(p); found {
		f.small.Remove(p)
	}
}

type findXSumPair struct {
	num  int
	freq int
}

func findXSumPairComparator(a, b findXSumPair) int {
	if a.freq != b.freq {
		return a.freq - b.freq
	}
	return a.num - b.num
}
