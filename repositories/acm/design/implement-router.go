package design

import (
	"sort"
)

type Packet struct {
	Source      int
	Destination int
	Timestamp   int
}

type Pair struct {
	Timestamps []int
}

type Router struct {
	memoryLimit      int
	packetQ          []Packet
	packetSet        map[Packet]struct{}
	destToTimestamps map[int]*Pair
}

func RouterConstructor(memoryLimit int) Router {
	return Router{
		memoryLimit:      memoryLimit,
		packetQ:          []Packet{},
		packetSet:        map[Packet]struct{}{},
		destToTimestamps: map[int]*Pair{},
	}
}

func (this *Router) AddPacket(source int, destination int, timestamp int) bool {
	p := Packet{source, destination, timestamp}
	if _, ok := this.packetSet[p]; ok {
		return false
	}
	if len(this.packetQ) == this.memoryLimit {
		this.ForwardPacket()
	}
	this.packetQ = append(this.packetQ, p)
	this.packetSet[p] = struct{}{}
	if _, ok := this.destToTimestamps[destination]; !ok {
		this.destToTimestamps[destination] = &Pair{}
	}
	this.destToTimestamps[destination].Timestamps = append(this.destToTimestamps[destination].Timestamps, timestamp)
	return true
}

func (this *Router) ForwardPacket() []int {
	if len(this.packetQ) == 0 {
		return []int{}
	}
	packet := this.packetQ[0]
	this.packetQ = this.packetQ[1:]
	delete(this.packetSet, packet)
	this.destToTimestamps[packet.Destination].Timestamps = this.destToTimestamps[packet.Destination].Timestamps[1:]
	return []int{packet.Source, packet.Destination, packet.Timestamp}
}

func (this *Router) GetCount(destination int, startTime int, endTime int) int {
	p, ok := this.destToTimestamps[destination]
	if !ok {
		return 0
	}
	left := sort.Search(len(p.Timestamps), func(i int) bool {
		return p.Timestamps[i] >= startTime
	})
	right := sort.Search(len(p.Timestamps), func(i int) bool {
		return p.Timestamps[i] > endTime
	})

	return right - left
}
