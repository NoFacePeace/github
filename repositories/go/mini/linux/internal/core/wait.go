package core

import (
	"sync"
	"unsafe"
)

const (
	WaitQueueFlagPriority = 0x10
)

// 等待队列头
type WaitQueueHead struct {
	sync.Mutex
	Head ListHead
}

// 等待队列元素
type WaitQueueEntry struct {
	Entry ListHead
	Func  WaitQueueFunc
	// 标志位
	Flags uint
}

// 等待队列元素执行函数
type WaitQueueFunc func(*WaitQueueEntry)

func InitWaitQueueFuncEntry(entry *WaitQueueEntry, f WaitQueueFunc) {
	entry.Func = f
}

func AddWaitQueue(head *WaitQueueHead, entry *WaitQueueEntry) {
	head.Lock()
	addWaitQueue(head, entry)
	defer head.Unlock()
}

func addWaitQueue(wqh *WaitQueueHead, wqe *WaitQueueEntry) {
	head := &wqh.Head
	var wq *WaitQueueEntry
	for wq = listFirstEntry(&wqh.Head); wq.Entry != wqh.Head; wq = ListNextEntry(wq) {
		if !(wq.Flags&WaitQueueFlagPriority != 0) {
			break
		}
		head = &wq.Entry
	}
	ListAdd(&wqe.Entry, head)
}

func listFirstEntry(lh *ListHead) *WaitQueueEntry {
	return (*WaitQueueEntry)(ContainerOf(unsafe.Pointer(lh), unsafe.Offsetof(WaitQueueEntry{}.Entry)))
}

func ListNextEntry(wqe *WaitQueueEntry) *WaitQueueEntry {
	next := wqe.Entry.Next
	return (*WaitQueueEntry)(ContainerOf(unsafe.Pointer(next), unsafe.Offsetof(WaitQueueEntry{}.Entry)))
}

func RemoveWaitQueue(wqh *WaitQueueHead, wqe *WaitQueueEntry) {
	wqh.Lock()
	defer wqh.Unlock()
	removeWaitQueue(wqh, wqe)
}

func removeWaitQueue(wqh *WaitQueueHead, wqe *WaitQueueEntry) {
	ListDel(&wqe.Entry)
}
