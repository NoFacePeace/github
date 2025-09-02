package fs

import "syscall"

type FdSet struct {
	syscall.FdSet
}

type FdSetBits struct {
}

type file struct{}

func (f *file) poll() {
}
