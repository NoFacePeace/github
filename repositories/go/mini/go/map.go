package mini

import (
	"sync/atomic"
	"unsafe"
)

// MapGroupSlots 是 Go 1.28 runtime 中一个 group 的 slot 数量。
const MapGroupSlots = 8

// MapCtrlEmpty 是一个 group 初始状态下的 control word。
const MapCtrlEmpty uint64 = 0x8080808080808080

const (
	mapMaxTableCapacity = 1024
	mapAverageLoadNum   = 7
	mapAverageLoadDen   = 8
)

var mapSeed uint64

// Map 模拟 Go 1.28 internal/runtime/maps.Map。
//
// 字段顺序与 Go 1.28 源码中的 Map 保持一致：
// src/internal/runtime/maps/map.go。
//
// Map 只是 map 的顶层描述符，不直接保存所有 key/value。
// 正常 map 会通过 dirPtr 找到 directory，再找到 table 和 group。
// 小 map 的 dirPtr 可以直接指向一个 group。
type Map struct {
	// used 是所有 table 中有效元素的数量，不包含 deleted slot。
	used uint64

	// seed 是当前 map 的随机哈希种子。
	seed uintptr

	// dirPtr 指向 directory；小 map 时也可以直接指向一个 group。
	dirPtr unsafe.Pointer

	// dirLen 是 directory 的长度。
	dirLen int

	// globalDepth 是 directory 查找使用的哈希位数。
	globalDepth uint8

	// globalShift 是 directory 查找时的哈希移位量。
	globalShift uint8

	// writing 用于检测并发写入。
	writing uint8

	// tombstonePossible 表示 map 中是否可能存在 deleted slot。
	tombstonePossible bool

	// clearSeq 用于迭代器检测 map 是否执行过 clear。
	clearSeq uint64
}

// MapTable 模拟 Go runtime 的 table。
//
// 一个 table 是完整的 Swiss Table。Map 可以通过 directory 指向一个或多个
// table；table 扩容时会被替换，超过最大容量后会 split 成两个 table。
type MapTable struct {
	// used 是当前 table 中有效元素的数量。
	used uint16

	// capacity 是当前 table 的 slot 总数。
	capacity uint16

	// growthLeft 是触发 rehash 前还可以使用的 slot 数量。
	growthLeft uint16

	// localDepth 是当前 table 使用的哈希位数。
	localDepth uint8

	// index 是 table 在 directory 中的起始下标。
	index int

	// groups 指向当前 table 的 group 数组。
	groups MapGroupsReference
}

// MapGroupsReference 模拟 runtime 的 groupsReference。
type MapGroupsReference struct {
	// data 指向 group 数组。
	data unsafe.Pointer

	// lengthMask 等于 group 数量减一，要求 group 数量是 2 的幂。
	lengthMask uint64
}

// MapGroup 模拟一个 group 的布局。
//
// Go runtime 的 group 根据编译配置可能使用两种布局：
// control word、keys 数组、elems 数组；
// 或 control word、交错排列的 key/elem slot。
// 这里采用易读的分离数组布局。
type MapGroup struct {
	// ctrls 的每个字节对应一个 slot。
	ctrls uint64

	// keys 保存 group 中的 key。
	keys [MapGroupSlots]any

	// elems 保存 group 中的 value。
	elems [MapGroupSlots]any
}

// NewMap 创建一个 Map 描述符。
//
// escape=false 模拟编译器判断 map 不逃逸的路径，使用结构体值创建：
//
//	m := Map{}
//
// escape=true 模拟 map 逃逸的路径，使用 new(Map) 创建堆上的描述符。
//
// escape 参数是对编译器逃逸分析结果的显式模拟；真实 Go 程序不会在 runtime
// 中通过布尔值决定逃逸，逃逸结论是在编译期间由编译器确定的。
func NewMap(hint int, escape bool) *Map {
	var m *Map
	var stackGroup *MapGroup
	if escape {
		m = new(Map)
	} else {
		stackMap := Map{}
		m = &stackMap
	}

	if hint < 0 {
		hint = 0
	}

	m.seed = nextMapSeed()
	if hint <= MapGroupSlots {
		if !escape {
			// 编译器对不逃逸的小 map 会预先准备一个 group，并写入 dirPtr。
			// 这里用局部结构体模拟编译器生成的栈上 group；由于示例返回
			// *Map，真实 Go 编译器可能让它随指针一起逃逸，这一点无法用
			// 普通函数完全复刻。
			stackGroup = &MapGroup{ctrls: MapCtrlEmpty}
			m.dirPtr = unsafe.Pointer(stackGroup)
		}
		return m
	}

	// 预估满足 hint 的 slot 数量。真实 runtime 使用平均装载因子计算目标容量。
	targetCapacity := (hint*mapAverageLoadDen + mapAverageLoadNum - 1) / mapAverageLoadNum
	if targetCapacity < hint {
		return m
	}

	directorySize := (targetCapacity + mapMaxTableCapacity - 1) / mapMaxTableCapacity
	directorySize = roundUpPowerOfTwo(directorySize)
	if directorySize <= 0 {
		return m
	}

	m.globalDepth = uint8(log2(directorySize))
	m.globalShift = uint8(uintSizeBits() - int(m.globalDepth))

	directory := make([]*MapTable, directorySize)
	tableCapacity := targetCapacity / directorySize
	for i := range directory {
		directory[i] = newMapTable(tableCapacity, i, m.globalDepth)
	}

	// directory 的底层数组由 dirPtr 指向；unsafe.Pointer 仍然会被 GC 识别为指针。
	m.dirPtr = unsafe.Pointer(&directory[0])
	m.dirLen = len(directory)
	return m
}

// nextMapSeed 模拟 runtime 为每个新 map 生成独立 seed 的过程。
func nextMapSeed() uintptr {
	return uintptr(atomic.AddUint64(&mapSeed, 1))
}

// newMapTable 创建一个 table 和它的 group 数组。
func newMapTable(capacity, index int, localDepth uint8) *MapTable {
	if capacity < MapGroupSlots {
		capacity = MapGroupSlots
	}
	capacity = roundUpPowerOfTwo(capacity)
	groupCount := capacity / MapGroupSlots
	groups := make([]MapGroup, groupCount)
	for i := range groups {
		groups[i].ctrls = MapCtrlEmpty
	}

	return &MapTable{
		capacity:   uint16(capacity),
		growthLeft: uint16(capacity * mapAverageLoadNum / mapAverageLoadDen),
		localDepth: localDepth,
		index:      index,
		groups: MapGroupsReference{
			data:       unsafe.Pointer(&groups[0]),
			lengthMask: uint64(groupCount - 1),
		},
	}
}

// roundUpPowerOfTwo 将正整数向上取整到 2 的幂。
func roundUpPowerOfTwo(value int) int {
	if value <= 1 {
		return 1
	}
	result := 1
	for result < value {
		if result > int(^uint(0)>>1)/2 {
			return 0
		}
		result <<= 1
	}
	return result
}

// log2 返回正的 2 的幂对应的指数。
func log2(value int) int {
	result := 0
	for value > 1 {
		value >>= 1
		result++
	}
	return result
}

// uintSizeBits 返回当前平台 uintptr 的位数。
func uintSizeBits() int {
	return int(unsafe.Sizeof(uintptr(0)) * 8)
}
