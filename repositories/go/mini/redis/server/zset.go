package server

import (
	"math"
	"math/rand"
)

const (
	zSkipListMaxLevel = 32
	zSkipListP        = 4
)

type zSkipList struct {
	header *zSkipListNode
	tail   *zSkipListNode
	length uint
	level  int
}

type zSkipListNode struct {
	ele      string
	score    float64
	backward *zSkipListNode
	level    []zSkipListLevel
}

type zSkipListLevel struct {
	forward *zSkipListNode
	// 到下个节点的跨越节点数目
	span uint
}

func zskCreate() *zSkipList {
	zsl := &zSkipList{}
	zsl.level = 1
	zsl.length = 0
	zsl.header = zslCreateNode(zSkipListMaxLevel, 0, "")
	for i := 0; i < zSkipListMaxLevel; i++ {
		zsl.header.level[i].forward = nil
		zsl.header.level[i].span = 0
	}
	zsl.header.backward = nil
	zsl.tail = nil
	return zsl
}

func zslCreateNode(level int, score float64, ele string) *zSkipListNode {
	zn := &zSkipListNode{}
	zn.score = 0
	zn.ele = ele
	zn.level = make([]zSkipListLevel, level)
	return zn
}

func zslInsert(zsl *zSkipList, score float64, ele string) *zSkipListNode {
	// 头节点
	x := zsl.header
	// 各层的前节点
	update := make([]*zSkipListNode, zSkipListMaxLevel)
	// 各层前节点的排名
	rank := make([]uint, zSkipListMaxLevel)
	// 计算各层前节点以及排名
	for i := zsl.level - 1; i >= 0; i-- {
		// 最高层，排名为零
		if i == zsl.level-1 {
			rank[i] = 0
		} else {
			// 其他层继承上一层的排名
			rank[i] = rank[i+1]
		}
		// 如果后节点分数低，前进
		for x.level[i].forward != nil && (x.level[i].forward.score < score || (x.level[i].forward.score == score && x.level[i].forward.ele < ele)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		// 记录对应层前节点
		update[i] = x
	}
	// 随机层数
	level := zslRandomLevel()
	if level > zsl.level {
		zsl.level = level
	}
	// 创建节点
	x = zslCreateNode(level, score, ele)
	// 添加到链表中
	for i := 0; i < level; i++ {
		// 当前节点的后节点是前节点的后节点
		x.level[i].forward = update[i].level[i].forward
		// 前节点的后节点是当前节点
		update[i].level[i].forward = x
		// 当前节点的跨越节点数目是前节点跨越节点数目减去（当前节点的排名 - 前节点排名）
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		// 前节点的跨越节点数目等于
		update[i].level[i].span = rank[0] - rank[i] + 1
	}
	// 未遍历的层跨越节点数目加一
	for i := level; i < zsl.level; i++ {
		update[i].level[i].span++
	}
	// 设置前节点指针
	if update[0] != zsl.header {
		x.backward = update[0]
	}
	// 设置后节点
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}
	// 长度加一
	zsl.length++
	return x
}

func zslRandomLevel() int {
	threshold := math.MaxInt / zSkipListP
	level := 1
	for rand.Intn(math.MaxInt) < threshold {
		level++
	}
	if level < zSkipListMaxLevel {
		return level
	}
	return zSkipListMaxLevel
}

func zslDelete(zsl *zSkipList, score float64, ele string) *zSkipListNode {
	x := zsl.header
	update := make([]*zSkipListNode, zSkipListMaxLevel)
	for i := zsl.level - 1; i >= 0; i-- {
		for isForward(&(x.level[i]), score, ele) {
			x = x.level[i].forward
		}
		update[i] = x
	}
	x = x.level[0].forward
	if x != nil && x.score == score && x.ele == ele {
		zslDeleteNode(zsl, x, update)
	}
	return x
}

func isForward(zl *zSkipListLevel, score float64, ele string) bool {
	if zl.forward == nil {
		return false
	}
	if zl.forward.score < score {
		return true
	}
	if zl.forward.score > score {
		return false
	}
	if zl.forward.ele < ele {
		return true
	}
	return false
}

func zslDeleteNode(zsl *zSkipList, x *zSkipListNode, update []*zSkipListNode) {
	for i := 0; i < zsl.level; i++ {
		if update[i].level[i].forward != x {
			update[i].level[i].span--
			continue
		}
		update[i].level[i].span += x.level[i].span - 1
		update[i].level[i].forward = x.level[i].forward
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	for zsl.level > 1 && zsl.header.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	zsl.length--
}
