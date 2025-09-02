package core

type PollWaitQueues struct {
	PT      PollTable
	Entries []PollTableEntry
}

type PollTable struct {
	qproc PollQueueProc
}

type PollTableEntry struct {
	File        *File
	Wait        WaitQueueEntry
	WaitAddress *WaitQueueHead
}

type PollQueueProc func(*File, *WaitQueueHead, *PollTable)

func InitPollFuncPtr(pt *PollTable, qproc PollQueueProc) {
	pt.qproc = qproc
}
