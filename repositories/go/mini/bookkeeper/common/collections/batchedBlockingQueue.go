package collections

type BatchedBlockingQueue interface {
}

type BatchedArrayBlockingQueue struct {
}

func NewBatchedArrayBlockingQueue() *BatchedArrayBlockingQueue {
	return &BatchedArrayBlockingQueue{}
}

type BlockingMpscQueue struct {
}

func NewBlockingMpscQueue() *BlockingMpscQueue {
	return &BlockingMpscQueue{}
}
