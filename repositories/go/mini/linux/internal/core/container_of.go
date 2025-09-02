package core

import "unsafe"

func ContainerOf(ptr unsafe.Pointer, fieldOffset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) - fieldOffset)
}
