package fs

import (
	"time"
	"unsafe"

	"github.com/NoFacePeace/github/repositories/go/mini/linux/internal/core"
)

// Select 阻塞直到超时或者有 IO 事件发生
func Select(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout time.Duration) (int, error) {
	return kernSelect(nfd, r, w, e, timeout)
}

func kernSelect(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout time.Duration) (int, error) {
	return coreSysSelect(nfd, r, w, e, timeout)
}

func coreSysSelect(nfd int, r *FdSet, w *FdSet, e *FdSet, timeout time.Duration) (int, error) {
	return doSelect(nfd, &FdSetBits{}, timeout)
}

func doSelect(n int, fds *FdSetBits, timeout time.Duration) (int, error) {
	table := &core.PollWaitQueues{}
	// 初始化 poll 等待队列
	pollInitWait(table)
	// for {
	// 	for i := 0; i < n; i++ {

	// 	}
	// }
	pollFreeWait(table)
	return 0, nil
}

// pollInitWait 初始化 Poll 等待队列
func pollInitWait(pwq *core.PollWaitQueues) {
	// 设置 Poll 内部等待队列的入队函数
	core.InitPollFuncPtr(&pwq.PT, pollWait)
}

// pollWait 等待队列入队函数，在 Poll 函数内部执行
func pollWait(f *core.File, address *core.WaitQueueHead, pt *core.PollTable) {
	// 获取 Poll 等待队列
	pwq := (*core.PollWaitQueues)(core.ContainerOf(unsafe.Pointer(pt), unsafe.Offsetof(core.PollWaitQueues{}.PT)))
	// 获取一个空的条目
	entry := pollGetEntry(pwq)
	// 绑定条目与 Poll 内部自身的等待队列
	entry.WaitAddress = address
	// 设置 Poll 实例自身等待队列的条目唤醒时的执行函数
	core.InitWaitQueueFuncEntry(&entry.Wait, pollWake)
	// 将 Poll 等待队列条目的等待实例
	core.AddWaitQueue(address, &entry.Wait)
}

func pollGetEntry(pwq *core.PollWaitQueues) *core.PollTableEntry {
	entry := core.PollTableEntry{}
	pwq.Entries = append(pwq.Entries, entry)
	return &entry
}

func pollWake(wait *core.WaitQueueEntry) {

}

func pollFreeWait(pwq *core.PollWaitQueues) {
	for i := 0; i < len(pwq.Entries); i++ {
		pollFreeEntry(&pwq.Entries[i])
	}
}

func pollFreeEntry(entry *core.PollTableEntry) {
	core.RemoveWaitQueue(entry.WaitAddress, &entry.Wait)
}
