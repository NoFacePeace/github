package mini

import (
	"testing"
	"unsafe"
)

func TestNewMapStackPath(t *testing.T) {
	m := NewMap(0, false)
	if m == nil {
		t.Fatal("NewMap returned nil")
	}
	if m.seed == 0 {
		t.Fatal("NewMap did not initialize seed")
	}
	if m.dirPtr == nil {
		t.Fatal("non-escaping small map should initialize group")
	}
	if m.dirLen != 0 {
		t.Fatalf("small map dirLen = %d, want 0", m.dirLen)
	}
}

func TestNewMapHeapPath(t *testing.T) {
	m := NewMap(0, true)
	if m.seed == 0 {
		t.Fatal("NewMap did not initialize seed")
	}
	if m.dirPtr != nil {
		t.Fatal("small map should defer group allocation")
	}
}

func TestNewMapWithHint(t *testing.T) {
	m := NewMap(100, true)
	if m.dirPtr == nil {
		t.Fatal("NewMap did not allocate directory")
	}
	if m.dirLen != 1 {
		t.Fatalf("NewMap dirLen = %d, want 1", m.dirLen)
	}

	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	if directory[0] == nil {
		t.Fatal("directory[0] is nil")
	}
	if directory[0].capacity < 100 {
		t.Fatalf("table capacity = %d, want at least 100", directory[0].capacity)
	}
	if directory[0].groups.data == nil {
		t.Fatal("table groups are not initialized")
	}
}
