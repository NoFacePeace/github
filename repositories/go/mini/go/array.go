package mini

import "unsafe"

// ArrayLength 表示示例数组的固定元素数量。
const ArrayLength = 3

// Array 模拟 [3]int 数组在运行时的值布局。
//
// Go 数组没有运行时头部，数组值本身就是连续存储的元素。
// 这个定义只是为了在代码中表达 [3]int 的实际数据布局。
type Array struct {
	// elements 是数组实际持有的连续元素。
	elements [ArrayLength]int
}

// CompileArray 模拟编译器内部的数组类型。
//
// Elem 使用 any 代替编译器内部的 Type 指针，Bound 表示数组的固定长度。
// 该结构只描述类型，不保存数组实例的数据。
type CompileArray struct {
	// Elem 表示数组元素的编译期类型。
	Elem any
	// Bound 表示数组的固定元素数量。
	Bound int64
}

// ArrayType 模拟 runtime 的 internal/abi.ArrayType。
//
// Type、Elem、Slice 和 Len 描述数组类型本身，这里使用 any 简化类型元数据。
// ArrayType 是类型元数据，不保存某个数组实例的数据地址。
// 普通数组访问直接使用编译期已知的长度和元素大小；只有接口、反射等需要在运行时识别
// 数组类型和长度时，编译器才会生成对这类类型元数据的引用。
type ArrayType struct {
	// Type 指向公共 runtime 类型元数据。
	Type any
	// Elem 指向数组元素的 runtime 类型元数据。
	Elem any
	// Slice 指向相同元素类型的切片类型。
	Slice any
	// Len 表示数组的固定元素数量。
	Len uintptr
}

// EmptyInterface 模拟 runtime 的空接口布局。
//
// Type 保存动态类型元数据，Data 保存具体值的数据地址。
// 数组放入 any 或 reflect.ValueOf 时，可以通过这两个字段识别数组类型并访问数据。
// 在这种场景下，编译器会把类型描述符地址和数据地址编译成机器码，运行时再填入接口值；
// 普通数组变量本身不会因为存在 ArrayType 而额外携带类型和长度字段。
type EmptyInterface struct {
	// Type 指向动态值的类型元数据。
	Type any
	// Data 指向动态值的数据地址。
	Data unsafe.Pointer
}
