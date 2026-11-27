package runner

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	bpfLoadWord = 0x20
	bpfEqual    = 0x15
	bpfTestBits = 0x45
	bpfReturn   = 0x06

	seccompAllow  = 0x7fff0000
	seccompKill   = 0x80000000
	seccompEPERM  = 0x00050001
	seccompENOSYS = 0x00050026
)

// 系统调用编号取自 Linux 的 amd64 和 arm64 ABI。
var bwrapDeniedSyscalls = [...]struct {
	amd64   uint32
	arm64   uint32
	verdict uint32
}{
	{246, 104, seccompEPERM}, // kexec_load
	{320, 294, seccompEPERM}, // kexec_file_load
	{175, 105, seccompEPERM}, // init_module
	{313, 273, seccompEPERM}, // finit_module
	{176, 106, seccompEPERM}, // delete_module
	{167, 224, seccompEPERM}, // swapon
	{168, 225, seccompEPERM}, // swapoff
	{321, 280, seccompEPERM}, // bpf
	{298, 241, seccompEPERM}, // perf_event_open
	{323, 282, seccompEPERM}, // userfaultfd
	{304, 265, seccompEPERM}, // open_by_handle_at
	// 让提供替代 I/O 路径的程序在 io_uring 不可用时回退。
	{425, 425, seccompENOSYS}, // io_uring_setup
	{426, 426, seccompENOSYS}, // io_uring_enter
	{427, 427, seccompENOSYS}, // io_uring_register
}

// buildBwrapSeccomp 生成 Bubblewrap 可读的经典 BPF 指令。
func buildBwrapSeccomp(arch string) ([]byte, error) {
	var auditArch uint32
	switch arch {
	case "amd64":
		auditArch = 0xc000003e
	case "arm64":
		auditArch = 0xc00000b7
	default:
		return nil, fmt.Errorf("unsupported seccomp architecture %q", arch)
	}

	var filter []byte
	emit := func(code uint16, yes, no byte, value uint32) {
		filter = binary.LittleEndian.AppendUint16(filter, code)
		filter = append(filter, yes, no)
		filter = binary.LittleEndian.AppendUint32(filter, value)
	}
	deny := func(number, verdict uint32) {
		emit(bpfEqual, 0, 1, number)
		emit(bpfReturn, 0, 0, verdict)
	}

	// 按 seccomp_data.arch 校验 ABI，再读取 seccomp_data.nr。
	emit(bpfLoadWord, 0, 0, 4)
	emit(bpfEqual, 1, 0, auditArch)
	emit(bpfReturn, 0, 0, seccompKill)
	emit(bpfLoadWord, 0, 0, 0)
	if arch == "amd64" {
		emit(bpfTestBits, 0, 1, 0x40000000)
		emit(bpfReturn, 0, 0, seccompKill)
		// 旧内核也可能在未设置 x32 标志位时接受此范围。
		for number := uint32(512); number <= 547; number++ {
			deny(number, seccompKill)
		}
	}
	for _, rule := range bwrapDeniedSyscalls {
		number := rule.arm64
		if arch == "amd64" {
			number = rule.amd64
		}
		deny(number, rule.verdict)
	}
	emit(bpfReturn, 0, 0, seccompAllow)
	return filter, nil
}

func (r *BwrapRuntime) writeSeccompFile() (string, error) {
	filter, err := buildBwrapSeccomp(runtime.GOARCH)
	if err != nil {
		return "", err
	}
	dir, err := filepath.Abs(filepath.Join(r.DataRoot, "processes"))
	if err != nil {
		return "", fmt.Errorf("resolve seccomp directory: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create seccomp directory: %w", err)
	}
	file, err := os.CreateTemp(dir, "execution-*.seccomp")
	if err != nil {
		return "", fmt.Errorf("create seccomp filter: %w", err)
	}
	_, writeErr := file.Write(filter)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return "", fmt.Errorf("write seccomp filter: %w", errors.Join(err, os.Remove(file.Name())))
	}
	return file.Name(), nil
}
