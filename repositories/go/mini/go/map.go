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
	mapControlDeleted   = 0xfe
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

// MapAccessInt 模拟 Go runtime 的 mapaccess 查找路径。
//
// 这里使用 int 作为示例 key/value 类型，避免引入真实 runtime 的 MapType、
// Hasher 和 Equal 函数。查找流程仍然对应 Go 1.28 的 getWithKey：
//
// 1. 计算 hash，并拆分为 H1 和 H2；
// 2. 小 map 直接检查 group；
// 3. 普通 map 先通过 directory 找 table；
// 4. 按探测序列检查 group；
// 5. 使用 H2 筛选候选 slot，再比较完整 key。
func MapAccessInt(m *Map, key int) (int, bool) {
	if m == nil || m.used == 0 || m.dirPtr == nil {
		return 0, false
	}

	hash := mapHashInt(key, m.seed)
	if m.dirLen == 0 {
		return mapAccessGroupInt((*MapGroup)(m.dirPtr), hash, key)
	}

	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	tableIndex := mapDirectoryIndex(m, hash)
	return mapAccessTableInt(directory[tableIndex], hash, key)
}

// MapDeleteInt 模拟 Go runtime 的 mapdelete 路径。
//
// 如果 group 中仍有 empty slot，删除后可以直接恢复为 empty；如果 group 已满，
// 则必须写入 deleted tombstone，避免截断其他 key 的探测路径。
func MapDeleteInt(m *Map, key int) bool {
	if m == nil || m.used == 0 || m.dirPtr == nil {
		return false
	}

	hash := mapHashInt(key, m.seed)
	if m.dirLen == 0 {
		group := (*MapGroup)(m.dirPtr)
		return mapDeleteGroupInt(m, group, hash, key)
	}

	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	tableIndex := mapDirectoryIndex(m, hash)
	return mapDeleteTableInt(m, directory[tableIndex], hash, key)
}

// MapInsertInt 模拟 Go runtime 的 mapassign 写入路径。
//
// 当前复刻只支持 int key/value，但保留了 runtime 的主要流程：
//
// 1. 小 map 直接写入 group；
// 2. group 写满后转换为 table；
// 3. table 没有 growthLeft 时执行 grow 或 split；
// 4. 新 key 写入 slot，已有 key 只覆盖 value。
func MapInsertInt(m *Map, key, value int) {
	if m == nil {
		panic("mini: assignment to nil map")
	}

	if m.dirPtr == nil {
		group := &MapGroup{ctrls: MapCtrlEmpty}
		m.dirPtr = unsafe.Pointer(group)
	}

	hash := mapHashInt(key, m.seed)
	if m.dirLen == 0 {
		group := (*MapGroup)(m.dirPtr)
		if done, inserted := mapInsertGroupInt(group, hash, key, value); done {
			if inserted {
				m.used++
			}
			return
		}

		// small group 满后，runtime 会创建 table 并迁移已有元素。
		mapGrowSmallInt(m)
	}

	for {
		directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
		tableIndex := mapDirectoryIndex(m, hash)
		table := directory[tableIndex]
		if mapInsertTableInt(m, table, hash, key, value) {
			return
		}

		// table 已经达到装载上限，扩容后重新定位并重试当前 key。
		mapGrowTableInt(m, tableIndex, table)
	}
}

// mapInsertGroupInt 尝试在单个 group 中写入 key/value。
//
// 第一个返回值表示操作是否完成，第二个返回值表示是否插入了新 key。
func mapInsertGroupInt(group *MapGroup, hash uintptr, key, value int) (bool, bool) {
	for i := range MapGroupSlots {
		control := mapControlByte(group.ctrls, i)
		if mapControlFull(control) && group.keys[i].(int) == key {
			group.elems[i] = value
			return true, false
		}
		if mapControlEmpty(control) {
			group.keys[i] = key
			group.elems[i] = value
			group.ctrls = mapSetControlByte(group.ctrls, i, mapControlFromH2(mapH2(hash)))
			return true, true
		}
	}
	return false, false
}

// mapInsertTableInt 尝试在 table 中写入 key/value。
//
// table 没有可用容量时返回 false，由调用方执行 grow 或 split。
func mapInsertTableInt(m *Map, table *MapTable, hash uintptr, key, value int) bool {
	sequence := mapMakeProbeSequence(mapH1(hash), table.groups.lengthMask)
	var firstDeletedGroup *MapGroup
	var firstDeletedSlot int
	for {
		group := mapGroupAt(table.groups, sequence.offset)
		for i := range MapGroupSlots {
			control := mapControlByte(group.ctrls, i)
			if mapControlFull(control) && group.keys[i].(int) == key {
				group.elems[i] = value
				return true
			}
			if control == mapControlDeleted && firstDeletedGroup == nil {
				firstDeletedGroup = group
				firstDeletedSlot = i
			}
			if mapControlEmpty(control) {
				targetGroup := group
				targetSlot := i
				if firstDeletedGroup != nil {
					targetGroup = firstDeletedGroup
					targetSlot = firstDeletedSlot
				}
				if table.growthLeft == 0 && firstDeletedGroup == nil {
					if !mapPruneTombstonesInt(table, m) {
						return false
					}
					return mapInsertTableInt(m, table, hash, key, value)
				}
				targetGroup.keys[targetSlot] = key
				targetGroup.elems[targetSlot] = value
				targetGroup.ctrls = mapSetControlByte(targetGroup.ctrls, targetSlot, mapControlFromH2(mapH2(hash)))
				table.used++
				if firstDeletedGroup == nil {
					table.growthLeft--
				}
				m.used++
				return true
			}
		}
		sequence = mapNextProbeSequence(sequence)
	}
}

// mapDeleteGroupInt 删除 small map 中的 key。
func mapDeleteGroupInt(m *Map, group *MapGroup, hash uintptr, key int) bool {
	matches := mapMatchH2(group.ctrls, mapH2(hash))
	for matches != 0 {
		slot := mapFirstMatch(&matches)
		if group.keys[slot].(int) != key {
			continue
		}
		group.keys[slot] = nil
		group.elems[slot] = nil
		group.ctrls = mapSetControlByte(group.ctrls, slot, mapControlEmptyValue())
		m.used--
		return true
	}
	return false
}

// mapDeleteTableInt 删除 table 中的 key，并按 group 状态决定 empty 或 tombstone。
func mapDeleteTableInt(m *Map, table *MapTable, hash uintptr, key int) bool {
	sequence := mapMakeProbeSequence(mapH1(hash), table.groups.lengthMask)
	for {
		group := mapGroupAt(table.groups, sequence.offset)
		matches := mapMatchH2(group.ctrls, mapH2(hash))
		for matches != 0 {
			slot := mapFirstMatch(&matches)
			if group.keys[slot].(int) != key {
				continue
			}

			group.keys[slot] = nil
			group.elems[slot] = nil
			table.used--
			m.used--
			if mapMatchEmpty(group.ctrls) != 0 {
				group.ctrls = mapSetControlByte(group.ctrls, slot, mapControlEmptyValue())
				table.growthLeft++
			} else {
				group.ctrls = mapSetControlByte(group.ctrls, slot, mapControlDeleted)
				m.tombstonePossible = true
			}
			return true
		}
		if mapMatchEmpty(group.ctrls) != 0 {
			return false
		}
		sequence = mapNextProbeSequence(sequence)
	}
}

// mapPruneTombstonesInt 清理不再位于有效 key 探测路径上的 tombstone。
//
// 真实 runtime 使用 bitset 记录必须保留的 group；这里使用 bool 切片表达相同
// 逻辑。清理只修改 control byte，不移动 key/value，避免破坏迭代语义。
func mapPruneTombstonesInt(table *MapTable, m *Map) bool {
	tombstones := mapCountTombstones(table)
	if tombstones*10 < int(table.capacity) {
		return false
	}

	groupCount := int(table.groups.lengthMask + 1)
	needed := make([]bool, groupCount)
	for groupIndex := range groupCount {
		group := mapGroupAt(table.groups, uint64(groupIndex))
		if mapMatchEmpty(group.ctrls) != 0 {
			needed[groupIndex] = true
		}

		for slot := range MapGroupSlots {
			if !mapControlFull(mapControlByte(group.ctrls, slot)) {
				continue
			}
			key := group.keys[slot].(int)
			hash := mapHashInt(key, m.seed)
			sequence := mapMakeProbeSequence(mapH1(hash), table.groups.lengthMask)
			for sequence.offset != uint64(groupIndex) {
				probeGroup := mapGroupAt(table.groups, sequence.offset)
				if mapMatchDeleted(probeGroup.ctrls) != 0 {
					needed[sequence.offset] = true
				}
				sequence = mapNextProbeSequence(sequence)
			}
		}
	}

	removable := 0
	for groupIndex := range groupCount {
		if needed[groupIndex] {
			continue
		}
		group := mapGroupAt(table.groups, uint64(groupIndex))
		removable += mapCountBits(mapMatchDeleted(group.ctrls))
	}
	if removable*10 < int(table.capacity) {
		return false
	}

	for groupIndex := range groupCount {
		if needed[groupIndex] {
			continue
		}
		group := mapGroupAt(table.groups, uint64(groupIndex))
		deleted := mapMatchDeleted(group.ctrls)
		for deleted != 0 {
			slot := mapFirstMatch(&deleted)
			group.ctrls = mapSetControlByte(group.ctrls, slot, mapControlEmptyValue())
			table.growthLeft++
		}
	}
	m.tombstonePossible = mapCountTombstones(table) != 0
	return true
}

func mapCountTombstones(table *MapTable) int {
	limit := int(table.capacity) * mapAverageLoadNum / mapAverageLoadDen
	return limit - int(table.used) - int(table.growthLeft)
}

// mapGrowSmallInt 将 small group 转换为第一个 table。
func mapGrowSmallInt(m *Map) {
	oldGroup := (*MapGroup)(m.dirPtr)
	table := newMapTable(MapGroupSlots*2, 0, 0)
	for i := range MapGroupSlots {
		control := mapControlByte(oldGroup.ctrls, i)
		if !mapControlFull(control) {
			continue
		}
		key := oldGroup.keys[i].(int)
		mapInsertKnownInt(table, mapHashInt(key, m.seed), key, oldGroup.elems[i].(int))
	}

	directory := make([]*MapTable, 1)
	directory[0] = table
	m.dirPtr = unsafe.Pointer(&directory[0])
	m.dirLen = 1
	m.globalDepth = 0
	m.globalShift = uint8(uintSizeBits())
}

// mapGrowTableInt 执行 table 翻倍或 split。
func mapGrowTableInt(m *Map, tableIndex int, oldTable *MapTable) {
	if int(oldTable.capacity)*2 <= mapMaxTableCapacity {
		newTable := newMapTable(int(oldTable.capacity)*2, oldTable.index, oldTable.localDepth)
		mapReinsertTableInt(newTable, oldTable, m.seed, 0)
		mapReplaceTableInt(m, oldTable, newTable)
		return
	}

	mapSplitTableInt(m, tableIndex, oldTable)
}

// mapReinsertTableInt 将旧 table 中的有效 slot 重新插入新 table。
func mapReinsertTableInt(target, source *MapTable, seed uintptr, splitMask uintptr) {
	for groupIndex := range int(source.groups.lengthMask + 1) {
		group := mapGroupAt(source.groups, uint64(groupIndex))
		for slot := range MapGroupSlots {
			control := mapControlByte(group.ctrls, slot)
			if !mapControlFull(control) {
				continue
			}
			key := group.keys[slot].(int)
			if splitMask != 0 && mapHashInt(key, seed)&splitMask != 0 {
				continue
			}
			mapInsertKnownInt(target, mapHashInt(key, seed), key, group.elems[slot].(int))
		}
	}
}

// mapInsertKnownInt 将已确认不存在的 key 放入新 table。
func mapInsertKnownInt(table *MapTable, hash uintptr, key, value int) {
	sequence := mapMakeProbeSequence(mapH1(hash), table.groups.lengthMask)
	for {
		group := mapGroupAt(table.groups, sequence.offset)
		for slot := range MapGroupSlots {
			if mapControlEmpty(mapControlByte(group.ctrls, slot)) {
				group.keys[slot] = key
				group.elems[slot] = value
				group.ctrls = mapSetControlByte(group.ctrls, slot, mapControlFromH2(mapH2(hash)))
				table.used++
				table.growthLeft--
				return
			}
		}
		sequence = mapNextProbeSequence(sequence)
	}
}

// mapReplaceTableInt 将 directory 中指向旧 table 的位置改为新 table。
func mapReplaceTableInt(m *Map, oldTable, newTable *MapTable) {
	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	for index := range directory {
		if directory[index] == oldTable {
			directory[index] = newTable
		}
	}
}

// mapSplitTableInt 将一个达到最大容量的 table 拆成两个 table。
func mapSplitTableInt(m *Map, tableIndex int, oldTable *MapTable) {
	newDepth := oldTable.localDepth + 1
	if oldTable.localDepth == m.globalDepth {
		oldDirectory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
		newDirectory := make([]*MapTable, len(oldDirectory)*2)
		for index, table := range oldDirectory {
			newDirectory[index*2] = table
			newDirectory[index*2+1] = table
		}
		m.globalDepth++
		m.globalShift--
		m.dirPtr = unsafe.Pointer(&newDirectory[0])
		m.dirLen = len(newDirectory)
	}

	left := newMapTable(mapMaxTableCapacity, tableIndex, newDepth)
	right := newMapTable(mapMaxTableCapacity, tableIndex+1, newDepth)
	splitMask := uintptr(1) << (uintSizeBits() - int(newDepth))
	mapReinsertTableInt(left, oldTable, m.seed, splitMask)

	for groupIndex := range int(oldTable.groups.lengthMask + 1) {
		group := mapGroupAt(oldTable.groups, uint64(groupIndex))
		for slot := range MapGroupSlots {
			if !mapControlFull(mapControlByte(group.ctrls, slot)) {
				continue
			}
			key := group.keys[slot].(int)
			if mapHashInt(key, m.seed)&splitMask != 0 {
				mapInsertKnownInt(right, mapHashInt(key, m.seed), key, group.elems[slot].(int))
			}
		}
	}

	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	for index := range directory {
		if directory[index] != oldTable {
			continue
		}
		if index&1 == 0 {
			directory[index] = left
		} else {
			directory[index] = right
		}
	}
}

// mapAccessGroupInt 查询一个 group。
func mapAccessGroupInt(group *MapGroup, hash uintptr, key int) (int, bool) {
	matches := mapMatchH2(group.ctrls, mapH2(hash))
	for matches != 0 {
		slot := mapFirstMatch(&matches)
		if group.keys[slot].(int) == key {
			return group.elems[slot].(int), true
		}
	}
	return 0, false
}

// mapAccessTableInt 按 Go runtime 的探测序列查询 table。
func mapAccessTableInt(table *MapTable, hash uintptr, key int) (int, bool) {
	sequence := mapMakeProbeSequence(mapH1(hash), table.groups.lengthMask)
	for {
		group := mapGroupAt(table.groups, sequence.offset)
		matches := mapMatchH2(group.ctrls, mapH2(hash))
		for matches != 0 {
			slot := mapFirstMatch(&matches)
			if group.keys[slot].(int) == key {
				return group.elems[slot].(int), true
			}
		}

		// 遇到 empty 才能结束查找；deleted 必须继续沿探测序列查找。
		if mapMatchEmpty(group.ctrls) != 0 {
			return 0, false
		}
		sequence = mapNextProbeSequence(sequence)
	}
}

type mapProbeSequence struct {
	mask   uint64
	offset uint64
	index  uint64
}

// mapMakeProbeSequence 对应 runtime 的 makeProbeSeq。
func mapMakeProbeSequence(hash uintptr, mask uint64) mapProbeSequence {
	return mapProbeSequence{
		mask:   mask,
		offset: uint64(hash) & mask,
	}
}

// mapNextProbeSequence 对应 runtime 的 probeSeq.next。
func mapNextProbeSequence(sequence mapProbeSequence) mapProbeSequence {
	sequence.index++
	sequence.offset = (sequence.offset + sequence.index) & sequence.mask
	return sequence
}

// mapGroupAt 根据 groupsReference 取得 group。
func mapGroupAt(groups MapGroupsReference, index uint64) *MapGroup {
	return &unsafe.Slice((*MapGroup)(groups.data), int(groups.lengthMask+1))[index]
}

// mapDirectoryIndex 对应 runtime 的 Map.directoryIndex。
func mapDirectoryIndex(m *Map, hash uintptr) int {
	if m.dirLen == 1 {
		return 0
	}
	return int(hash >> (m.globalShift & 63))
}

// mapHashInt 模拟 runtime 针对 int key 的类型专用 Hasher。
func mapHashInt(key int, seed uintptr) uintptr {
	hash := uint64(uintptr(key)) + uint64(seed)*0x9e3779b97f4a7c15
	hash ^= hash >> 30
	hash *= 0xbf58476d1ce4e5b9
	hash ^= hash >> 27
	hash *= 0x94d049bb133111eb
	hash ^= hash >> 31
	return uintptr(hash)
}

func mapH1(hash uintptr) uintptr {
	return hash >> 7
}

func mapH2(hash uintptr) uint8 {
	return uint8(hash & 0x7f)
}

// mapControlByte 读取 control word 中第 index 个 control byte。
func mapControlByte(control uint64, index int) uint8 {
	return uint8(control >> (uint(index) * 8))
}

func mapSetControlByte(control uint64, index int, value uint8) uint64 {
	mask := uint64(0xff) << (uint(index) * 8)
	return (control &^ mask) | uint64(value)<<(uint(index)*8)
}

func mapControlFromH2(hash uint8) uint8 {
	return hash & 0x7f
}

func mapControlEmptyValue() uint8 {
	return 0x80
}

func mapControlFull(control uint8) bool {
	return control&0x80 == 0
}

func mapControlEmpty(control uint8) bool {
	return control == mapControlEmptyValue()
}

// mapMatchH2 返回匹配 H2 的 slot 位图。
//
// 真实 runtime 使用位运算或 SIMD 一次匹配 8 个 control byte；这里用循环
// 表达相同语义，便于直接观察每个 slot 的判断过程。
func mapMatchH2(control uint64, hash uint8) uint8 {
	var matches uint8
	for i := range MapGroupSlots {
		value := mapControlByte(control, i)
		if mapControlFull(value) && value&0x7f == hash {
			matches |= 1 << i
		}
	}
	return matches
}

func mapMatchEmpty(control uint64) uint8 {
	var matches uint8
	for i := range MapGroupSlots {
		if mapControlEmpty(mapControlByte(control, i)) {
			matches |= 1 << i
		}
	}
	return matches
}

func mapMatchDeleted(control uint64) uint8 {
	var matches uint8
	for i := range MapGroupSlots {
		if mapControlByte(control, i) == mapControlDeleted {
			matches |= 1 << i
		}
	}
	return matches
}

func mapCountBits(value uint8) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

func mapFirstMatch(matches *uint8) int {
	for i := range MapGroupSlots {
		if *matches&(1<<i) != 0 {
			*matches &^= 1 << i
			return i
		}
	}
	return -1
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
