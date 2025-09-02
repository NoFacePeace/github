package core

import "sync"

// 等待队列头
type WaitQueueHead struct {
	sync.Mutex
}

// 等待队列元素
type WaitQueueEntry struct {
	Entry ListHead
	Func  WaitQueueFunc
}

// 等待队列元素执行函数
type WaitQueueFunc func(*WaitQueueEntry)

func InitWaitQueueFuncEntry(entry *WaitQueueEntry, f WaitQueueFunc) {
	entry.Func = f
}

func AddWaitQueue(head *WaitQueueHead, entry *WaitQueueEntry) {
	head.Lock()
}
