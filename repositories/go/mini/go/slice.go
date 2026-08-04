package mini

import "unsafe"

const (
	// sliceVariableMakeStackBytes 是变量容量 make 的默认小栈缓冲区大小。
	sliceVariableMakeStackBytes = 32
	// sliceImplicitStackBytes 是编译器允许隐式栈变量使用的默认大小。
	sliceImplicitStackBytes = 64 * 1024
	// sliceIntBytes 是当前平台 int 的字节数。
	sliceIntBytes = int(unsafe.Sizeof(int(0)))
	// sliceVariableMakeIntCapacity 是 32 字节最多容纳的 int 数量。
	sliceVariableMakeIntCapacity = sliceVariableMakeStackBytes / sliceIntBytes
	// sliceImplicitStackIntCapacity 是 64 KB 最多容纳的 int 数量。
	sliceImplicitStackIntCapacity = sliceImplicitStackBytes / sliceIntBytes
)

// Slice 模拟 Go runtime 中切片描述符的底层布局。
//
// runtime.slice 的定义包含一个底层数组指针、当前长度和容量。
// Slice 只描述切片头部，不保存底层数组本身，也不能替代 Go 原生切片。
type Slice struct {
	// array 指向底层数组的首元素。
	array unsafe.Pointer
	// len 表示切片当前包含的元素数量。
	len int
	// cap 表示从首元素开始到底层数组末尾的容量。
	cap int
}

// NewSlice 模拟编译器和 runtime 一起创建 []int 的过程。
//
// 编译期间，编译器会先进行逃逸分析，并判断底层数组是否足够小。
// 变量容量 make 的默认小栈缓冲区是 32 字节，因此当前 64 位环境下
// 4 个 int 可能走小栈缓冲区。容量是编译期常量时，编译器还可能使用
// 最大 64 KB 的隐式栈数组。
//
// 运行期间，如果容量超过栈分配条件，则使用 make 分配底层数组。
// 真实 Go 代码中，逃逸分析由编译器根据调用关系完成；本函数只是用注释
// 表达这个过程，没有额外传入 escapes 参数。
func NewSlice(length, capacity int) Slice {
	validateSliceArguments(length, capacity)

	var source []int
	if capacity <= sliceVariableMakeIntCapacity {
		// 编译器可能将这个小的固定数组放在栈上。
		source = []int{0, 0, 0, 0}
		source = source[:length:capacity]
	} else if capacity <= sliceImplicitStackIntCapacity {
		// 容量是编译期常量且不逃逸时，编译器可能生成类似
		// var arr [cap]int; source = arr[:length:capacity] 的代码。
		var stackBacking [sliceImplicitStackIntCapacity]int
		source = stackBacking[:length:capacity]
	} else {
		// runtime 在运行期间按 length 和 capacity 分配堆上的底层数组。
		source = make([]int, length, capacity)
	}

	return Slice{
		array: unsafe.Pointer(unsafe.SliceData(source)),
		len:   len(source),
		cap:   cap(source),
	}
}

// AppendSlice 模拟编译器展开 append 和 runtime.growslice 的过程。
//
// values 表示本次 append 追加的元素。num 由 len(values) 得到：
// 普通 append 对应参数个数，append(source, other...) 对应 len(other)。
// 容量足够时复用原底层数组；容量不足时按 runtime 的扩容策略分配新数组、
// 复制旧元素和新增元素，并返回新的切片描述符。
func AppendSlice(source Slice, values []int) Slice {
	num := len(values)
	if source.len > int(^uint(0)>>1)-num {
		panic("growslice: len out of range")
	}

	newLen := source.len + num
	// 编译器生成的快速路径：容量足够时不分配新数组。
	if newLen <= source.cap {
		target := unsafe.Slice((*int)(source.array), source.cap)
		copy(target[source.len:newLen], values)
		source.len = newLen
		return source
	}

	oldCap := source.cap
	doubleCap := oldCap + oldCap
	newCap := oldCap
	if newLen > doubleCap {
		newCap = newLen
	} else if oldCap < 256 {
		newCap = doubleCap
	} else {
		// runtime 对大切片从翻倍平滑过渡到约 1.25 倍增长。
		for {
			newCap += (newCap + 3*256) >> 2
			if uint(newCap) >= uint(newLen) {
				break
			}
		}
		if newCap <= 0 {
			newCap = newLen
		}
	}

	// runtime.growslice 在运行期间分配新的底层数组并复制旧元素。
	oldData := unsafe.Slice((*int)(source.array), source.len)
	newData := make([]int, newCap)
	copy(newData, oldData)
	copy(newData[source.len:newLen], values)

	return Slice{
		array: unsafe.Pointer(unsafe.SliceData(newData)),
		len:   newLen,
		cap:   newCap,
	}
}

// validateSliceArguments 校验模拟 makeslice 使用的长度、容量和内存大小。
func validateSliceArguments(length, capacity int) {
	if length < 0 {
		panic("makeslice: len out of range")
	}
	if capacity < 0 {
		panic("makeslice: cap out of range")
	}
	if length > capacity {
		panic("makeslice: len out of range")
	}
}
