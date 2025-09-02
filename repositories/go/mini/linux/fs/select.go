package fs

import (
	"time"
	"unsafe"

	"github.com/NoFacePeace/github/repositories/go/mini/linux/internal/core"
)

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
	pollInitWait(table)
	for {
		for i := 0; i < n; i++ {

		}
	}
	return 0, nil
}

func pollInitWait(pwq *core.PollWaitQueues) {
	core.InitPollFuncPtr(&pwq.PT, pollWait)
}

func pollWait(f *core.File, address *core.WaitQueueHead, pt *core.PollTable) {
	pwq := (*core.PollWaitQueues)(core.ContainerOf(unsafe.Pointer(pt), unsafe.Offsetof(core.PollWaitQueues{}.PT)))
	entry := pollGetEntry(pwq)
	entry.WaitAddress = address
	core.InitWaitQueueFuncEntry(&entry.Wait, pollWake)
}

func pollGetEntry(pwq *core.PollWaitQueues) *core.PollTableEntry {
	entry := core.PollTableEntry{}
	pwq.Entries = append(pwq.Entries, entry)
	return &entry
}

func pollWake(*core.WaitQueueEntry) {

}
