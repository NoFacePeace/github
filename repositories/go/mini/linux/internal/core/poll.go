package core

// PollWaitQueues 支持 Poll 接口等待队列
type PollWaitQueues struct {
	//
	PT PollTable
	// Entries 等待队列条目数组
	Entries []PollTableEntry
}

type PollTable struct {
	qproc PollQueueProc
}

// Poll 等待队列元素
type PollTableEntry struct {
	File        *File
	Wait        WaitQueueEntry
	WaitAddress *WaitQueueHead
}

type PollQueueProc func(*File, *WaitQueueHead, *PollTable)

// InitPollFuncPtr 设置 File 等待队列入队函数
func InitPollFuncPtr(pt *PollTable, qproc PollQueueProc) {
	pt.qproc = qproc
}
