package mini

import "unsafe"

// tmpStringBufferSize 与 Go 运行时临时字符串缓冲区的大小保持一致。
const tmpStringBufferSize = 64

// String 模拟 Go 运行时中 string 的底层布局。
type String struct {
	// str 指向字符串底层字节数据。
	str unsafe.Pointer
	// len 表示字符串的字节长度。
	len int
}

// NewString 模拟 Go 创建字符串描述符的过程。
//
// 对于字符串字面量，编译器会将字节数据放入只读数据区，并构造包含数据指针和字节长度的
// string 描述符。该函数读取 source 的底层数据指针与长度，不分配内存也不复制字节。
// 需要从可修改字节切片创建独立字符串时，应使用 BytesToString。
func NewString(source string) String {
	return String{
		str: unsafe.Pointer(unsafe.StringData(source)),
		len: len(source),
	}
}

// ConcatStrings 模拟 Go 的字符串拼接流程。
//
// Go 编译器会将 + 表达式转换为 runtime.concatstring2 到
// runtime.concatstring5 或 runtime.concatstrings 调用。运行时会跳过空字符串、
// 计算总字节长度；没有非空字符串时返回空字符串，只有一个非空字符串时在安全条件下
// 直接复用该字符串。其余情况调用 rawstringtmp 分配连续空间，再依次复制各字符串。
//
// 拼接结果不逃逸且长度不超过 64 字节时，编译器会向 runtime 传入调用方栈上的临时
// 缓冲区；否则 runtime 通过 mallocgc 在 GC 堆上分配。当前示例使用相同的 64 字节
// 阈值，但普通 Go 函数返回 string 时无法完全复刻 runtime 的栈缓冲调用约定。
func ConcatStrings(parts ...string) string {
	// last 记录最后一个非空字符串的位置，count 记录非空字符串数量。
	last := 0
	total := 0
	count := 0

	for i, part := range parts {
		// 空字符串不会改变拼接结果，无需参与后续复制。
		if len(part) == 0 {
			continue
		}
		// 防止长度累加溢出。
		if total+len(part) < total {
			panic("string concatenation too long")
		}
		total += len(part)
		last = i
		count++
	}

	// 所有输入均为空时，返回空字符串。
	if count == 0 {
		return ""
	}
	// 只有一个非空字符串时可以直接复用，无需分配和复制。
	if count == 1 {
		return parts[last]
	}

	// 小结果优先写入固定大小的临时缓冲区，模拟运行时的 64 字节阈值。
	// 当前函数返回 string，转换后的字符串必须在函数返回后有效，通常会逃逸到堆。
	// runtime 的栈缓冲优化依赖编译器内部调用约定，普通 Go 函数无法完全复刻。
	if total <= tmpStringBufferSize {
		var stackBuffer [tmpStringBufferSize]byte
		offset := 0
		for _, part := range parts {
			offset += copy(stackBuffer[offset:], part)
		}
		return string(stackBuffer[:total])
	}

	// 大结果使用动态缓冲区，并将各字符串按顺序复制进去。
	buffer := make([]byte, total)
	offset := 0
	for _, part := range parts {
		offset += copy(buffer[offset:], part)
	}
	// 将填充完成的字节缓冲区转换为字符串。
	return string(buffer)
}

// BytesToString 模拟 runtime.slicebytetostring 的主要转换流程。
//
// 空切片直接转换为空字符串；不超过 64 字节时使用临时缓冲区，否则创建动态缓冲区。
// 两条路径都会先复制源字节，再使用 unsafe.String 构造只读字符串，因此后续修改 source
// 不会影响结果。runtime 对单字节字符串使用静态查找表，并由编译器处理不需复制的特殊
// 短生命周期场景；普通 Go 函数无法安全、完整地复刻这些内部优化。临时缓冲区是否最终
// 留在栈上由编译器逃逸分析决定。
func BytesToString(source []byte) string {
	if len(source) == 0 {
		return ""
	}

	if len(source) <= tmpStringBufferSize {
		var stackBuffer [tmpStringBufferSize]byte
		copy(stackBuffer[:], source)
		return unsafe.String(&stackBuffer[0], len(source))
	}

	heapBuffer := make([]byte, len(source))
	copy(heapBuffer, source)
	return unsafe.String(unsafe.SliceData(heapBuffer), len(heapBuffer))
}

// StringToBytes 模拟 runtime.stringtoslicebyte 的主要转换流程。
//
// 不超过 64 字节时使用临时缓冲区，否则创建动态缓冲区，随后复制字符串的全部字节。
// 复制使返回的字节切片可安全修改，而不会修改原字符串的底层数据。缓冲区是否最终留在
// 栈上由编译器逃逸分析决定。
func StringToBytes(source string) []byte {
	if len(source) <= tmpStringBufferSize {
		var stackBuffer [tmpStringBufferSize]byte
		copy(stackBuffer[:], source)
		return stackBuffer[:len(source)]
	}

	result := make([]byte, len(source))
	copy(result, source)
	return result
}
