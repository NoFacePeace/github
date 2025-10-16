package core

type ListHead struct {
	Prev *ListHead
	Next *ListHead
}

func ListAdd(new *ListHead, head *ListHead) {
	listAdd(new, head, head.Next)
}

func listAdd(new *ListHead, prev *ListHead, next *ListHead) {
	next.Prev = new
	new.Next = next
	new.Prev = prev
	prev.Next = new
}

func ListDel(entry *ListHead) {
	listDelEntry(entry)
	entry.Next = nil
	entry.Prev = nil
}

func listDelEntry(entry *ListHead) {
	listDel(entry.Prev, entry.Next)
}

func listDel(prev *ListHead, next *ListHead) {
	next.Prev = prev
	prev.Next = next
}
