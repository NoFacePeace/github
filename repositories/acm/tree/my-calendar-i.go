package tree

import "github.com/emirpasic/gods/trees/redblacktree"

type MyCalendar struct {
	tree *redblacktree.Tree
}

func Constructor() MyCalendar {
	tree := redblacktree.NewWithIntComparator()
	return MyCalendar{tree}
}

func (this *MyCalendar) Book(startTime int, endTime int) bool {
	if this.tree.Empty() {
		this.tree.Put(startTime, endTime)
		return true
	}
	node, found := this.tree.Floor(startTime)
	if found {
		end := node.Value.(int)
		if end > startTime {
			return false
		}
	}
	node, found = this.tree.Ceiling(startTime)
	if !found {
		this.tree.Put(startTime, endTime)
		return true
	}
	start := node.Key.(int)
	if start < endTime {
		return false
	}
	this.tree.Put(startTime, endTime)
	return true
}
