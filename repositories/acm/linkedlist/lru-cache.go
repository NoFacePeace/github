package linkedlist

type LRUCache struct {
	capacity int
	m        map[int]*LRUCacheNode
	head     *LRUCacheNode
}
type LRUCacheNode struct {
	left  *LRUCacheNode
	right *LRUCacheNode
	val   int
	key   int
}

func NewLRUCache(capacity int) LRUCache {
	head := &LRUCacheNode{}
	head.left = head
	head.right = head
	return LRUCache{
		head:     head,
		m:        map[int]*LRUCacheNode{},
		capacity: capacity,
	}
}

func (this *LRUCache) Get(key int) int {
	node, ok := this.m[key]
	if !ok {
		return -1
	}
	this.move(node)
	return node.val
}

func (this *LRUCache) Put(key int, value int) {
	node, ok := this.m[key]
	if ok {
		node.val = value
		this.move(node)
		return
	}
	node = &LRUCacheNode{
		val: value,
		key: key,
	}
	next := this.head.right
	next.left = node
	node.right = next
	node.left = this.head
	this.head.right = node
	this.m[key] = node
	if len(this.m) > this.capacity {
		last := this.head.left
		last.left.right = this.head
		this.head.left = last.left
		delete(this.m, last.key)
	}
}
func (this *LRUCache) move(node *LRUCacheNode) {
	left := node.left
	right := node.right
	left.right = right
	right.left = left
	next := this.head.right
	this.head.right = node
	node.left = this.head
	node.right = next
	next.left = node
}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
